//
// condor-launcher launches jobs on an HTCondor cluster.
//
// This service connects to an AMQP broker's "jobs" exchange and waits for
// messages sent with the "jobs.launches" key. It then turns the request
// into an iplant.cmd, config, job, and irods_config file in /tmp/<user>/<UUID>
// and calls out to condor_submit to submit the job to the cluster.
//
// Since it launches jobs by executing the condor_submit command it shouldn't
// run inside a Docker container. Our Condor cluster is moderately large and
// requires a lot of ports to be opened up, which doesn't play nicely with
// Docker.
//
// Required configuration keys are:
//   amqp.uri
//   irods.user
//   irods.pass
//   irods.host
//   irods.port
//   irods.base
//   irods.resc
//   irods.zone
//   condor.condor_config
//   condor.path_env_var
//   condor.log_path
//   condor.request_disk
//   porklock.image
//   porklock.tag
//
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"text/template"
	"time"

	"github.com/cyverse-de/configurate"
	"github.com/cyverse-de/go-events/ping"
	"github.com/cyverse-de/logcabin"
	"github.com/cyverse-de/messaging"
	"github.com/cyverse-de/model"
	"github.com/cyverse-de/version"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

var log = logrus.WithFields(logrus.Fields{
	"service": "condor-launcher",
	"art-id":  "condor-launcher",
	"group":   "org.cyverse",
})

func init() {
	var err error
	logrus.SetFormatter(&logrus.JSONFormatter{})
	SubmissionTemplate, err = template.New("condor_submit").Parse(SubmissionTemplateText)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse submission template text"))
	}

	JobConfigTemplate, err = template.New("job_config").Parse(JobConfigTemplateText)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse job config template text"))
	}

	IRODSConfigTemplate, err = template.New("irods_config").Parse(IRODSConfigTemplateText)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse irods config template text"))
	}
}

const pingKey = "events.condor-launcher.ping"
const pongKey = "events.condor-launcher.pong"

// Messenger defines an interface for handling AMQP operations. This is the
// subset of functionality needed by job-status-recorder.
type Messenger interface {
	AddConsumer(string, string, string, string, messaging.MessageHandler)
	Close()
	Listen()
	Publish(string, []byte) error
	SetupPublishing(string) error
	PublishJobUpdate(*messaging.UpdateMessage) error
}

// CondorLauncher contains the condor-launcher application state.
type CondorLauncher struct {
	cfg    *viper.Viper
	client Messenger
	fs     fsys
}

// New returns a new *CondorLauncher
func New(c *viper.Viper, client Messenger, fs fsys) *CondorLauncher {
	return &CondorLauncher{
		cfg:    c,
		client: client,
		fs:     fs,
	}
}

func (cl *CondorLauncher) submit(cmdPath string, s *model.Job) (string, error) {
	csPath, err := exec.LookPath("condor_submit")
	if err != nil {
		return "", errors.Wrapf(err, "failed to find condor_submit in $PATH")
	}

	if !path.IsAbs(csPath) {
		csPath, err = filepath.Abs(csPath)
		if err != nil {
			return "", errors.Wrapf(err, "failed to get the absolute path to %s", csPath)
		}
	}

	cmd := exec.Command(csPath, cmdPath)
	cmd.Dir = path.Dir(cmdPath)

	cmd.Env = []string{
		fmt.Sprintf("PATH=%s", cl.cfg.GetString("condor.path_env_var")),
		fmt.Sprintf("CONDOR_CONFIG=%s", cl.cfg.GetString("condor.condor_config")),
	}

	output, err := cmd.CombinedOutput()
	log.Infof("Output of condor_submit:\n%s\n", output)
	if err != nil {
		return "", errors.Wrapf(err, "failed to execute %s", csPath)
	}

	log.Infof("Extracted ID: %s\n", string(model.ExtractJobID(output)))

	return string(model.ExtractJobID(output)), err
}

func (cl *CondorLauncher) launch(s *model.Job) (string, error) {
	sdir, err := CreateSubmissionDirectory(s)
	if err != nil {
		log.Errorf("%+v\n", errors.Wrap(err, "failed to create submission directory"))
		return "", err
	}

	cmd, _, _, err := CreateSubmissionFiles(sdir, cl.cfg, s)
	if err != nil {
		log.Errorf("%+v\n", errors.Wrap(err, "failed to create submission files"))
		return "", err
	}

	id, err := cl.submit(cmd, s)
	if err != nil {
		log.Errorf("%+v\n", errors.Wrap(err, "failed to submit job"))
		return "", err
	}

	log.Infof("Condor job id is %s\n", id)

	return id, err
}

func (cl *CondorLauncher) stop(s *model.Job) (string, error) {
	crPath, err := exec.LookPath("condor_rm")
	log.Infof("condor_rm found at %s", crPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to find condor_rm on the $PATH")
	}

	if !path.IsAbs(crPath) {
		crPath, err = filepath.Abs(crPath)
		if err != nil {
			return "", errors.Wrapf(err, "failed to get the absolute path for %s", crPath)
		}
	}

	pathEnv := cl.cfg.GetString("condor.path_env_var")
	condorConfig := cl.cfg.GetString("condor.condor_config")

	cmd := exec.Command(crPath, s.CondorID)
	cmd.Env = []string{
		fmt.Sprintf("PATH=%s", pathEnv),
		fmt.Sprintf("CONDOR_CONFIG=%s", condorConfig),
	}

	output, err := cmd.CombinedOutput()
	log.Infof("condor_rm output for job %s:\n%s\n", s.CondorID, output)
	if err != nil {
		return "", errors.Wrapf(err, "failed to execute %s", crPath)
	}

	return string(output), err
}

// startHeldTicker starts up the code that periodically fires and clean up held
// jobs
func (cl *CondorLauncher) startHeldTicker(client *messaging.Client) (*time.Ticker, error) {
	d, err := time.ParseDuration("30s")
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse duration '30s'")
	}
	t := time.NewTicker(d)
	go func(t *time.Ticker, client *messaging.Client) {
		for {
			select {
			case <-t.C:
				cl.killHeldJobs(client)
			}
		}
	}(t, client)
	return t, nil
}

// handlePing is the handler for ping events.
func (cl *CondorLauncher) handlePing(delivery amqp.Delivery) {
	log.Infoln("Received ping")

	out, err := json.Marshal(&ping.Pong{})
	if err != nil {
		log.Errorf("%+v\n", errors.Wrap(err, "failed to marshal pong response"))
	}

	log.Infoln("Sent pong")

	if err = cl.client.Publish(pongKey, out); err != nil {
		log.Errorf("%+v\n", errors.Wrap(err, "failed to publish pong response"))
	}
}

// handleEvents accepts an amqp message, acks it, and delegates handling it to
// another function.
func (cl *CondorLauncher) handleEvents(delivery amqp.Delivery) {
	if err := delivery.Ack(false); err != nil {
		log.Errorf("%+v\n", errors.Wrap(err, "failed to ack amqp event delivery"))
	}

	switch delivery.RoutingKey {
	case pingKey:
		cl.handlePing(delivery)
	default:
		log.Errorf("%+v\n", fmt.Errorf("unhandled event with routing key of %s", delivery.RoutingKey))
	}
}

// handleLaunchRequests triggers Condor jobs in response to launch request messages.
func (cl *CondorLauncher) handleLaunchRequests(delivery amqp.Delivery) {
	body := delivery.Body

	if err := delivery.Ack(false); err != nil {
		log.Error(errors.Wrap(err, "failed to ack amqp launch request delivery"))
	}

	req := messaging.JobRequest{}
	err := json.Unmarshal(body, &req)
	if err != nil {
		log.Errorf("%+v\n", errors.Wrap(err, "failed to unmarshal launch request json"))
		log.Error(string(body[:]))
		return
	}

	if req.Job.RequestDisk == "" {
		req.Job.RequestDisk = "0"
	}

	switch req.Command {
	case messaging.Launch:
		jobID, err := cl.launch(req.Job)
		if err != nil {
			log.Errorf("%+v\n", err)
			err = cl.client.PublishJobUpdate(&messaging.UpdateMessage{
				Job:     req.Job,
				State:   messaging.FailedState,
				Message: fmt.Sprintf("condor-launcher failed to launch job:\n %s", err),
			})
			if err != nil {
				log.Errorf("%+v\n", errors.Wrap(err, "failed to publish launch failure job update"))
			}
		} else {
			log.Infof("Launched Condor ID %s", jobID)
			err = cl.client.PublishJobUpdate(&messaging.UpdateMessage{
				Job:     req.Job,
				State:   messaging.SubmittedState,
				Message: fmt.Sprintf("Launched Condor ID %s", jobID),
			})
			if err != nil {
				log.Errorf("%+v\n", errors.Wrap(err, "failed to publish successful launch job update"))
			}
		}
	}
}

func main() {
	var (
		cfgPath     = flag.String("config", "", "Path to the config file. Required.")
		showVersion = flag.Bool("version", false, "Print the version information")
		err         error
	)

	flag.Parse()

	logcabin.Init("condor-launcher", "condor-launcher")

	if *showVersion {
		version.AppVersion()
		os.Exit(0)
	}

	if *cfgPath == "" {
		fmt.Println("Error: --config must be set.")
		flag.PrintDefaults()
		os.Exit(-1)
	}

	cfg, err := configurate.InitDefaults(*cfgPath, configurate.JobServicesDefaults)
	if err != nil {
		log.Fatalf("%+v\n", errors.Wrap(err, "failed to initialize configuration defaults"))
	}
	log.Infoln("Done reading config.")

	uri := cfg.GetString("amqp.uri")
	exchangeName := cfg.GetString("amqp.exchange.name")
	exchangeType := cfg.GetString("amqp.exchange.type")

	client, err := messaging.NewClient(uri, true)
	if err != nil {
		log.Fatalf("%+v\n", errors.Wrap(err, "failed to create new AMQP client"))
	}
	defer client.Close()

	localfs := &osys{}
	launcher := New(cfg, client, localfs)

	launcher.client.SetupPublishing(exchangeName)

	go launcher.client.Listen()

	ticker, err := launcher.startHeldTicker(client)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
	log.Infof("Started up the held state ticker: %#v", ticker)

	launcher.RegisterStopHandler(client)

	launcher.client.AddConsumer(
		exchangeName,
		exchangeType,
		"condor_launcher_events",
		"events.condor-launcher.*",
		launcher.handleEvents,
	)

	// Accept and handle messages sent out with the jobs.launches routing key.
	launcher.client.AddConsumer(
		exchangeName,
		exchangeType,
		"condor_launches",
		messaging.LaunchesKey,
		launcher.handleLaunchRequests,
	)

	spin := make(chan int)
	<-spin
}

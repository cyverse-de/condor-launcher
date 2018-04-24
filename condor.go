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
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"text/template"
	"time"

	"github.com/cyverse-de/configurate"
	"github.com/cyverse-de/go-events/ping"
	"github.com/cyverse-de/logcabin"
	"github.com/cyverse-de/version"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/cyverse-de/messaging.v3"
	"gopkg.in/cyverse-de/model.v2"

	"github.com/cyverse-de/condor-launcher/jobs"
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
	IRODSConfigTemplate, err = template.New("irods_config").Parse(IRODSConfigTemplateText)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse irods config template text"))
	}
}

func ackDelivery(delivery amqp.Delivery, logMsgOnErr string) {
	if err := delivery.Ack(false); err != nil {
		log.Error(errors.Wrap(err, logMsgOnErr))
	}
}

func rejectDelivery(delivery amqp.Delivery, requeue bool, logMsgOnErr string) {
	if err := delivery.Reject(requeue); err != nil {
		log.Error(errors.Wrap(err, logMsgOnErr))
	}
}

const pingKey = "events.condor-launcher.ping"
const pongKey = "events.condor-launcher.pong"

// Messenger defines an interface for handling AMQP operations. This is the
// subset of functionality needed by job-status-recorder.
type Messenger interface {
	AddConsumer(string, string, string, string, messaging.MessageHandler, int)
	Close()
	Listen()
	Publish(string, []byte) error
	SetupPublishing(string) error
	PublishJobUpdate(*messaging.UpdateMessage) error
	DeleteQueue(name string) error
}

// CondorLauncher contains the condor-launcher application state.
type CondorLauncher struct {
	cfg          *viper.Viper
	client       Messenger
	fs           fsys
	v            VaultOperator
	cubbyMount   string // the path to where the cubbyhole backend is rooted in Vault
	condorSubmit string //path to the condor_submit executable
	condorRm     string // path to the condor_rm executable
}

// New returns a new *CondorLauncher
func New(c *viper.Viper, client Messenger, fs fsys, condorSubmit, condorRm string) *CondorLauncher {
	return &CondorLauncher{
		cfg:          c,
		client:       client,
		fs:           fs,
		condorSubmit: condorSubmit,
		condorRm:     condorRm,
	}
}

func (cl *CondorLauncher) storeConfig(s *model.Job) (string, error) {
	uselimit := len(s.Inputs()) + 2 // 2 comes from 1 for writing, one for the output job.

	childToken, err := cl.v.ChildToken(uselimit)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate child token")
	}
	log.Infof("generated a child token for job %s", s.InvocationID)

	cfgData := &IRODSConfig{
		IRODSHost: cl.cfg.GetString("irods.host"),
		IRODSPort: cl.cfg.GetString("irods.port"),
		IRODSUser: cl.cfg.GetString("irods.user"),
		IRODSPass: cl.cfg.GetString("irods.pass"),
		IRODSBase: cl.cfg.GetString("irods.base"),
		IRODSResc: cl.cfg.GetString("irods.resc"),
		IRODSZone: cl.cfg.GetString("irods.zone"),
	}
	fileContent, err := GenerateFile(IRODSConfigTemplate, cfgData)
	if err != nil {
		return "", err
	}
	log.Infof("generated the irods config for job %s", s.InvocationID)

	if err = cl.v.StoreConfig(
		childToken,
		cl.cfg.GetString("vault.irods.mount_path"),
		s.InvocationID,
		fileContent.Bytes(),
	); err != nil {
		return "", err
	}
	log.Infof("stored the irods config for job %s in vault", s.InvocationID)

	return childToken, nil
}

func (cl *CondorLauncher) launch(s *model.Job, condorPath, condorConfig string) (string, error) {

	// Ensure that the logs directory exists for the job.
	sdir := s.CondorLogDirectory()
	if path.Base(sdir) != "logs" {
		sdir = path.Join(sdir, "logs")
	}
	err := os.MkdirAll(sdir, 0755)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create the directory %s", sdir)
	}

	// Store the Condor configs in Vault.
	childToken, err := cl.storeConfig(s)
	if err != nil {
		return "", err
	}

	// Create a copy of the configuration that also contains the Vault child token.
	cfgCopy := CopyConfig(cl.cfg)
	cfgCopy.Set("vault.child_token.token", childToken)

	// Generate the submission files, always using the condor job submission format for now.
	jobSubmissionBuilder, err := jobs.NewJobSubmissionBuilder("condor", cfgCopy)
	if err != nil {
		return "", err
	}
	submissionPath, err := jobSubmissionBuilder.Build(s, sdir)
	if err != nil {
		return "", err
	}

	// Submit the job to Condor.
	cmd := exec.Command(cl.condorSubmit, submissionPath)
	cmd.Dir = path.Dir(submissionPath)
	cmd.Env = []string{
		fmt.Sprintf("PATH=%s", condorPath),
		fmt.Sprintf("CONDOR_CONFIG=%s", condorConfig),
	}
	output, err := cmd.CombinedOutput()
	log.Infof("Output of condor_submit:\n%s\n", output)
	if err != nil {
		return "", errors.Wrapf(err, "failed to execute %s", cl.condorSubmit)
	}

	// Log the Condor job ID.
	id := string(model.ExtractJobID(output))
	log.Infof("Condor job id is %s\n", id)

	return id, err
}

// handleEvents accepts an amqp message, acks it, and delegates handling it to
// another function.
func (cl *CondorLauncher) routeEvents(delivery amqp.Delivery) {
	ackDelivery(delivery, "failed to ack amqp event delivery")

	switch delivery.RoutingKey {
	case pingKey:
		log.Infoln("Received ping")
		out, err := json.Marshal(&ping.Pong{})
		if err != nil {
			log.Errorf("%+v\n", errors.Wrap(err, "failed to marshal pong response"))
		}
		log.Infoln("Sent pong")
		if err = cl.client.Publish(pongKey, out); err != nil {
			log.Errorf("%+v\n", errors.Wrap(err, "failed to publish pong response"))
		}
	default:
		log.Errorf("%+v\n", fmt.Errorf("unhandled event with routing key of %s", delivery.RoutingKey))
	}
}

// handleLaunchRequests triggers Condor jobs in response to launch request messages.
func (cl *CondorLauncher) handleLaunchRequests(condorPath, condorConfig string) func(d amqp.Delivery) {
	return func(delivery amqp.Delivery) {
		body := delivery.Body
		requeueOnErr := !delivery.Redelivered

		req := messaging.JobRequest{}
		err := json.Unmarshal(body, &req)
		if err != nil {
			log.Errorf("%+v\n", errors.Wrap(err, "failed to unmarshal launch request json"))
			log.Error(string(body[:]))

			rejectDelivery(delivery, requeueOnErr, "failed to Reject amqp Launch request delivery")

			return
		}

		if req.Job.RequestDisk == "" {
			req.Job.RequestDisk = "0"
		}

		switch req.Command {
		case messaging.Launch:
			jobID, err := cl.launch(req.Job, condorPath, condorConfig)
			if err != nil {
				log.Errorf("%+v\n", err)

				if !requeueOnErr {
					err = cl.client.PublishJobUpdate(&messaging.UpdateMessage{
						Job:     req.Job,
						State:   messaging.FailedState,
						Message: fmt.Sprintf("condor-launcher failed to launch job:\n %s", err),
					})
					if err != nil {
						log.Errorf("%+v\n", errors.Wrap(err, "failed to publish launch failure job update"))
					}
				}

				rejectDelivery(delivery, requeueOnErr, "failed to Reject amqp Launch request delivery")
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

				ackDelivery(delivery, "failed to ACK amqp Launch request delivery")
			}
		default:
			log.Errorf("condor_launches message handler got unrecognized command: %+v\n", req.Command)
			ackDelivery(delivery, "failed to ACK amqp Launch request delivery")
		}
	}
}

func (launcher *CondorLauncher) stopJob(invocationID, condorPath, condorConfig string) error {
	var (
		condorRMOutput []byte
		err            error
	)

	log.Infof("Running condor_rm for %s", invocationID)
	if condorRMOutput, err = ExecCondorRm(invocationID, condorPath, condorConfig); err != nil {
		log.Errorf("%+v\n", errors.Wrapf(err, "failed to run 'condor_rm %s'", invocationID))
		return err
	}

	fauxJob := model.New(launcher.cfg)
	fauxJob.InvocationID = invocationID
	update := &messaging.UpdateMessage{
		Job:     fauxJob,
		State:   messaging.FailedState,
		Message: "Job was killed",
	}
	if err = launcher.client.PublishJobUpdate(update); err != nil {
		log.Errorf("%+v\n", errors.Wrap(err, "failed to publish job update for a stopped job"))
	}
	log.Infof("condor_rm output for job %s:\n%s", invocationID, condorRMOutput)

	launcher.client.DeleteQueue(messaging.StopQueueName(invocationID))

	return nil
}

func (cl *CondorLauncher) stopHandler(condorPath, condorConfig string) func(d amqp.Delivery) {
	return func(d amqp.Delivery) {
		var (
			requeueOnErr bool
			invID        string
			err          error
		)

		requeueOnErr = !d.Redelivered

		stopRequest := &messaging.StopRequest{}
		if err = json.Unmarshal(d.Body, stopRequest); err != nil {
			log.Errorf("%+v\n", errors.Wrap(err, "failed to unmarshal the stop request body"))
			rejectDelivery(d, requeueOnErr, "failed to Reject StopRequest")
			return
		}

		invID = stopRequest.InvocationID

		if err = cl.stopJob(invID, condorPath, condorConfig); err != nil {
			rejectDelivery(d, requeueOnErr, fmt.Sprintf("failed to Reject StopRequest for %s", invID))
		} else {
			ackDelivery(d, fmt.Sprintf("failed to ACK StopRequest for %s", invID))
		}
	}
}

func killHeldJobs(launcher *CondorLauncher, condorPath, condorConfig string) {
	var (
		err         error
		cmdOutput   []byte
		heldEntries []string
	)
	log.Infoln("Looking for jobs in the held state...")
	if cmdOutput, err = ExecCondorQHeldIDs(condorPath, condorConfig); err != nil {
		log.Errorf("%+v\n", errors.Wrap(err, "error running condor_q"))
		return
	}
	heldEntries = heldQueueInvocationIDs(cmdOutput)
	log.Infof("There are %d jobs in the held state", len(heldEntries))
	for _, invocationID := range heldEntries {
		if invocationID != "" {
			log.Infof("Sending stop request for invocation id %s", invocationID)
			if err = launcher.stopJob(invocationID, condorPath, condorConfig); err != nil {
				log.Errorf("%+v\n", errors.Wrap(err, "error sending stop request"))
			}
		}
	}
}

// startHeldTicker starts up the code that periodically fires and clean up held
// jobs
func startHeldTicker(launcher *CondorLauncher, condorPath, condorConfig string) (*time.Ticker, error) {
	d, err := time.ParseDuration("30s")
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse duration '30s'")
	}
	t := time.NewTicker(d)
	go func(t *time.Ticker, launcher *CondorLauncher) {
		for {
			select {
			case <-t.C:
				killHeldJobs(launcher, condorPath, condorConfig)
			}
		}
	}(t, launcher)
	return t, nil
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

	csPath := findExecPath("condor_submit")
	crPath := findExecPath("condor_rm")

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

	launcher := New(cfg, client, &osys{}, csPath, crPath)
	launcher.client.SetupPublishing(exchangeName)
	go launcher.client.Listen()

	condorPath := cfg.GetString("condor.path_env_var")
	condorConfig := cfg.GetString("condor.condor_config")

	launcher.v, err = VaultInit(
		cfg.GetString("vault.token"),
		cfg.GetString("vault.url"),
	)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}

	if err = launcher.v.MountCubbyhole(cfg.GetString("vault.irods.mount_path")); err != nil {
		log.Fatalf("%+v\n", err)
	}

	ticker, err := startHeldTicker(
		launcher,
		condorPath,
		condorConfig,
	)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
	log.Infof("Started up the held state ticker: %#v", ticker)

	launcher.client.AddConsumer(
		exchangeName,
		exchangeType,
		"condor-launcher-stops",
		messaging.StopRequestKey("*"),
		launcher.stopHandler(condorPath, condorConfig),
		cfg.GetInt("amqp.prefetch.stops"),
	)

	launcher.client.AddConsumer(
		exchangeName,
		exchangeType,
		"condor_launcher_events",
		"events.condor-launcher.*",
		launcher.routeEvents,
		cfg.GetInt("amqp.prefetch.events"),
	)

	// Accept and handle messages sent out with the jobs.launches routing key.
	launcher.client.AddConsumer(
		exchangeName,
		exchangeType,
		"condor_launches",
		messaging.LaunchesKey,
		launcher.handleLaunchRequests(condorPath, condorConfig),
		cfg.GetInt("amqp.prefetch.launches"),
	)

	spin := make(chan int)
	<-spin
}

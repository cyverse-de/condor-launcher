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
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
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
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

const pingKey = "events.condor-launcher.ping"
const pongKey = "events.condor-launcher.pong"

// SubmissionTemplateText is the text of the template for the HTCondor
// submission file.
const SubmissionTemplateText = `universe = vanilla
executable = /usr/local/bin/road-runner
rank = mips{{ if .UsesVolumes }}
requirements = (HAS_HOST_MOUNTS == True){{ end }}
arguments = --config config --job job
output = script-output.log
error = script-error.log
log = condor.log{{if .Group}}
accounting_group = {{.Group}}
accounting_group_user = {{.Submitter}}{{end}}
request_disk = {{.RequestDisk}}
+IpcUuid = "{{.InvocationID}}"
+IpcJobId = "generated_script"
+IpcUsername = "{{.Submitter}}"
+IpcUserGroups = {{.FormatUserGroups}}
concurrency_limits = {{.UserIDForSubmission}}
{{with $x := index .Steps 0}}+IpcExe = "{{$x.Component.Name}}"{{end}}
{{with $x := index .Steps 0}}+IpcExePath = "{{$x.Component.Location}}"{{end}}
should_transfer_files = YES
transfer_input_files = irods-config,iplant.cmd,config,job
transfer_output_files = workingvolume/logs/logs-stdout-output,workingvolume/logs/logs-stderr-output
when_to_transfer_output = ON_EXIT_OR_EVICT
notification = NEVER
queue
`

// JobConfigTemplateText is the text of the template for the HTCondor submission
// file.
const JobConfigTemplateText = `amqp:
uri: {{.GetString "amqp.uri"}}
exchange:
	name: {{.GetString "amqp.exchange.name"}}
	type: {{.GetString "amqp.exchange.type"}}
irods:
base: "{{.GetString "irods.base"}}"
porklock:
image: "{{.GetString "porklock.image"}}"
tag: "{{.GetString "porklock.tag"}}"
condor:
filter_files: "{{.GetString "condor.filter_files"}}"`

// IRODSConfigTemplateText is the text of the template for porklock's iRODS
// config file.
const IRODSConfigTemplateText = `porklock.irods-host = {{.IRODSHost}}
porklock.irods-port = {{.IRODSPort}}
porklock.irods-user = {{.IRODSUser}}
porklock.irods-pass = {{.IRODSPass}}
porklock.irods-home = {{.IRODSBase}}
porklock.irods-zone = {{.IRODSZone}}
porklock.irods-resc = {{.IRODSResc}}
`

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

// fsys defines an interface for file operations.
type fsys interface {
	MkdirAll(string, os.FileMode) error
	WriteFile(string, []byte, os.FileMode) error
}

// osys is an implementation of fsys that hits the os and ioutil packages.
type osys struct{}

func (o *osys) MkdirAll(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}

func (o *osys) WriteFile(path string, contents []byte, mode os.FileMode) error {
	return ioutil.WriteFile(path, contents, mode)
}

// CondorLauncher contains the condor-launcher application state.
type CondorLauncher struct {
	cfg                 *viper.Viper
	client              Messenger
	fs                  fsys
	submissionTemplate  *template.Template
	jobConfigTemplate   *template.Template
	irodsConfigTemplate *template.Template
}

// New returns a new *CondorLauncher
func New(c *viper.Viper, client Messenger, fs fsys) (*CondorLauncher, error) {
	cl := &CondorLauncher{
		cfg:    c,
		client: client,
		fs:     fs,
	}
	st, err := template.New("condor_submit").Parse(SubmissionTemplateText)
	if err != nil {
		return cl, err
	}
	cl.submissionTemplate = st
	jct, err := template.New("job_config").Parse(JobConfigTemplateText)
	if err != nil {
		return cl, err
	}
	cl.jobConfigTemplate = jct
	ict, err := template.New("irods_config").Parse(IRODSConfigTemplateText)
	if err != nil {
		return cl, err
	}
	cl.irodsConfigTemplate = ict
	return cl, err
}

// GenerateCondorSubmit returns a string (or error) containing the contents
// of what should go into an HTCondor submission file.
func (cl *CondorLauncher) GenerateCondorSubmit(submission *model.Job) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	err := cl.submissionTemplate.Execute(&buffer, submission)
	return &buffer, err
}

type scriptable struct {
	model.Job
	DC []model.VolumesFrom
	CI []model.ContainerImage
}

// GenerateJobConfig creates a string containing the config that gets passed
// into the job.
func (cl *CondorLauncher) GenerateJobConfig() (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	err := cl.jobConfigTemplate.Execute(&buffer, cl.cfg)
	return &buffer, err
}

type irodsconfig struct {
	IRODSHost string
	IRODSPort string
	IRODSUser string
	IRODSPass string
	IRODSZone string
	IRODSBase string
	IRODSResc string
}

// GenerateIRODSConfig returns the contents of the irods-config file as a string.
func (cl *CondorLauncher) GenerateIRODSConfig() (*bytes.Buffer, error) {
	c := &irodsconfig{
		IRODSHost: cl.cfg.GetString("irods.host"),
		IRODSPort: cl.cfg.GetString("irods.port"),
		IRODSUser: cl.cfg.GetString("irods.user"),
		IRODSPass: cl.cfg.GetString("irods.pass"),
		IRODSBase: cl.cfg.GetString("irods.base"),
		IRODSResc: cl.cfg.GetString("irods.resc"),
		IRODSZone: cl.cfg.GetString("irods.zone"),
	}
	var buffer bytes.Buffer
	err := cl.irodsConfigTemplate.Execute(&buffer, c)
	return &buffer, err
}

// CreateSubmissionDirectory creates a directory for a submission and returns the path to it as a string.
func (cl *CondorLauncher) CreateSubmissionDirectory(s *model.Job) (string, error) {
	dirPath := s.CondorLogDirectory()
	if path.Base(dirPath) != "logs" {
		dirPath = path.Join(dirPath, "logs")
	}
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return "", err
	}
	return dirPath, err
}

// CreateSubmissionFiles creates the iplant.cmd file inside the
// directory designated by 'dir'. The return values are the path to the iplant.cmd
// file, and any errors, in that order.
func (cl *CondorLauncher) CreateSubmissionFiles(dir string, s *model.Job) (string, string, string, error) {
	cmdContents, err := cl.GenerateCondorSubmit(s)
	if err != nil {
		return "", "", "", err
	}

	jobConfigContents, err := cl.GenerateJobConfig()
	if err != nil {
		return "", "", "", err
	}

	jobContents, err := json.Marshal(s)
	if err != nil {
		return "", "", "", err
	}

	irodsContents, err := cl.GenerateIRODSConfig()
	if err != nil {
		return "", "", "", err
	}
	subfiles := []struct {
		filename    string
		filecontent []byte
		permissions os.FileMode
	}{
		{filename: path.Join(dir, "iplant.cmd"), filecontent: cmdContents.Bytes(), permissions: 0644},
		{filename: path.Join(dir, "config"), filecontent: jobConfigContents.Bytes(), permissions: 0644},
		{filename: path.Join(dir, "job"), filecontent: jobContents, permissions: 0644},
		{filename: path.Join(dir, "irods-config"), filecontent: irodsContents.Bytes(), permissions: 0644},
	}

	for _, sf := range subfiles {
		err = ioutil.WriteFile(sf.filename, sf.filecontent, sf.permissions)
		if err != nil {
			return "", "", "", err
		}
	}
	return subfiles[0].filename, subfiles[1].filename, subfiles[2].filename, nil
}

func (cl *CondorLauncher) submit(cmdPath string, s *model.Job) (string, error) {
	csPath, err := exec.LookPath("condor_submit")
	if err != nil {
		return "", err
	}

	if !path.IsAbs(csPath) {
		csPath, err = filepath.Abs(csPath)
		if err != nil {
			return "", err
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
		return "", err
	}

	log.Infof("Extracted ID: %s\n", string(model.ExtractJobID(output)))

	return string(model.ExtractJobID(output)), err
}

func (cl *CondorLauncher) launch(s *model.Job) (string, error) {
	sdir, err := cl.CreateSubmissionDirectory(s)
	if err != nil {
		log.Errorf("Error creating submission directory:\n%s\n", err)
		return "", err
	}

	cmd, _, _, err := cl.CreateSubmissionFiles(sdir, s)
	if err != nil {
		log.Errorf("Error creating submission files:\n%s", err)
		return "", err
	}

	id, err := cl.submit(cmd, s)
	if err != nil {
		log.Errorf("Error submitting job:\n%s", err)
		return "", err
	}

	log.Infof("Condor job id is %s\n", id)

	return id, err
}

func (cl *CondorLauncher) stop(s *model.Job) (string, error) {
	crPath, err := exec.LookPath("condor_rm")
	log.Infof("condor_rm found at %s", crPath)
	if err != nil {
		return "", err
	}

	if !path.IsAbs(crPath) {
		crPath, err = filepath.Abs(crPath)
		if err != nil {
			return "", err
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
		return "", err
	}

	return string(output), err
}

// startHeldTicker starts up the code that periodically fires and clean up held
// jobs
func (cl *CondorLauncher) startHeldTicker(client *messaging.Client) (*time.Ticker, error) {
	d, err := time.ParseDuration("30s")
	if err != nil {
		return nil, err
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
		log.Error(err)
	}

	log.Infoln("Sent pong")

	if err = cl.client.Publish(pongKey, out); err != nil {
		log.Error(err)
	}
}

// handleEvents accepts an amqp message, acks it, and delegates handling it to
// another function.
func (cl *CondorLauncher) handleEvents(delivery amqp.Delivery) {
	if err := delivery.Ack(false); err != nil {
		log.Error(err)
	}

	switch delivery.RoutingKey {
	case pingKey:
		cl.handlePing(delivery)
	default:
		log.Errorf("unhandled event with routing key of %s", delivery.RoutingKey)
	}
}

// handleLaunchRequests triggers Condor jobs in response to launch request messages.
func (cl *CondorLauncher) handleLaunchRequests(delivery amqp.Delivery) {
	body := delivery.Body

	if err := delivery.Ack(false); err != nil {
		log.Error(err)
	}

	req := messaging.JobRequest{}
	err := json.Unmarshal(body, &req)
	if err != nil {
		log.Error(err)
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
			log.Error(err)
			err = cl.client.PublishJobUpdate(&messaging.UpdateMessage{
				Job:     req.Job,
				State:   messaging.FailedState,
				Message: fmt.Sprintf("condor-launcher failed to launch job:\n %s", err),
			})
			if err != nil {
				log.Error(err)
			}
		} else {
			log.Infof("Launched Condor ID %s", jobID)
			err = cl.client.PublishJobUpdate(&messaging.UpdateMessage{
				Job:     req.Job,
				State:   messaging.SubmittedState,
				Message: fmt.Sprintf("Launched Condor ID %s", jobID),
			})
			if err != nil {
				log.Error(err)
			}
		}
	}
}

func main() {
	var (
		cfgPath     = flag.String("config", "", "Path to the config file. Required.")
		showVersion = flag.Bool("version", false, "Print the version information")
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
		log.Fatal(err)
	}
	log.Infoln("Done reading config.")

	uri := cfg.GetString("amqp.uri")
	exchangeName := cfg.GetString("amqp.exchange.name")
	exchangeType := cfg.GetString("amqp.exchange.type")

	client, err := messaging.NewClient(uri, true)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	localfs := &osys{}
	launcher, err := New(cfg, client, localfs)
	if err != nil {
		log.Fatal(err)
	}

	launcher.client.SetupPublishing(exchangeName)

	go launcher.client.Listen()

	ticker, err := launcher.startHeldTicker(client)
	if err != nil {
		log.Fatal(err)
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

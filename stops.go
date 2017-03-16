package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/cyverse-de/messaging"
	"github.com/cyverse-de/model"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/streadway/amqp"
)

// ExecCondorQ runs the condor_q -long command and returns its output.
func ExecCondorQ(cfg *viper.Viper) ([]byte, error) {
	var (
		output []byte
		err    error
	)

	csPath, err := exec.LookPath("condor_q")
	if err != nil {
		return output, errors.Wrap(err, "failed to find condor_q on the $PATH")
	}

	if !path.IsAbs(csPath) {
		csPath, err = filepath.Abs(csPath)
		if err != nil {
			return output, errors.Wrapf(err, "failed to get the absolute path of %s", csPath)
		}
	}

	cmd := exec.Command(csPath, "-long")

	pathEnv := cfg.GetString("condor.path_env_var")

	condorCfg := cfg.GetString("condor.condor_config")

	cmd.Env = []string{
		fmt.Sprintf("PATH=%s", pathEnv),
		fmt.Sprintf("CONDOR_CONFIG=%s", condorCfg),
	}

	output, err = cmd.CombinedOutput()
	if err != nil {
		return output, errors.Wrapf(err, "failed to get the output of the command '%s %s'", csPath, "-long")
	}

	return output, nil
}

// ExecCondorRm runs condor_rm, passing it the condor ID. Returns the output
// of the command and passibly an error.
func ExecCondorRm(condorID string, cfg *viper.Viper) ([]byte, error) {
	var (
		output []byte
		err    error
	)

	crPath, err := exec.LookPath("condor_rm")
	log.Infof("condor_rm found at %s", crPath)
	if err != nil {
		return output, errors.Wrap(err, "failed to find condor_rm on the $PATH")
	}

	if !path.IsAbs(crPath) {
		crPath, err = filepath.Abs(crPath)
		if err != nil {
			return output, errors.Wrapf(err, "failed to get the absolute path of %s", crPath)
		}
	}

	pathEnv := cfg.GetString("condor.path_env_var")

	condorConfig := cfg.GetString("condor.condor_config")

	cmd := exec.Command(crPath, condorID)
	cmd.Env = []string{
		fmt.Sprintf("PATH=%s", pathEnv),
		fmt.Sprintf("CONDOR_CONFIG=%s", condorConfig),
	}

	output, err = cmd.CombinedOutput()
	log.Infof("condor_rm output for job %s:\n%s\n", condorID, output)
	if err != nil {
		return output, errors.Wrapf(err, "failed to get the output of '%s %s'", crPath, condorID)
	}

	return output, nil
}

type queueEntry struct {
	CondorID     string
	InvocationID string
	IsHeld       bool
}

var (
	condorIDKey  = []byte("ClusterId")
	statusKey    = []byte("JobStatus")
	newLineBytes = []byte("\n")
	equalBytes   = []byte(" = ")
	ipcUUIDBytes = []byte("IpcUuid")
)

func queueEntries(output []byte) []queueEntry {
	var (
		retval   []queueEntry
		condorID []byte
		jobID    []byte
		statusID []byte
	)
	chunks := bytes.Split(output, []byte("\n\n"))
	for _, chunk := range chunks {
		lines := bytes.Split(chunk, newLineBytes)
		for _, line := range lines {
			if bytes.Contains(line, equalBytes) {
				lineChunks := bytes.Split(line, equalBytes)
				if len(lineChunks) >= 2 {
					key := lineChunks[0]
					value := lineChunks[1]
					switch {
					case bytes.Equal(key, condorIDKey):
						condorID = bytes.TrimSpace(value)
					case bytes.Equal(key, ipcUUIDBytes):
						jobID = bytes.TrimSpace(value)
					case bytes.Equal(key, statusKey):
						statusID = bytes.TrimSpace(value)
					}
				}
			}
		}
		newEntry := queueEntry{
			CondorID:     string(condorID),
			InvocationID: string(bytes.Trim(jobID, "\" ")),
			IsHeld:       bytes.Equal(statusID, []byte("5")),
		}
		retval = append(retval, newEntry)
	}
	return retval
}

// queueEntriesByInvocationID returns a list of queueEntries for the given
// invocation ID, which is stored in the "IpcUuid" field of the condor_q output.
// This function does not call condor_q, it should be passed the output of the
// condor_q command.
func queueEntriesByInvocationID(output []byte, invID string) []queueEntry {
	var retval []queueEntry
	entries := queueEntries(output)
	for _, entry := range entries {
		if entry.InvocationID == invID {
			retval = append(retval, entry)
		}
	}
	return retval
}

func heldQueueEntries(output []byte) []queueEntry {
	var retval []queueEntry
	entries := queueEntries(output)
	for _, entry := range entries {
		if entry.IsHeld {
			retval = append(retval, entry)
		}
	}
	return retval
}

func (cl *CondorLauncher) killHeldJobs(client *messaging.Client) {
	var (
		err         error
		cmdOutput   []byte
		heldEntries []queueEntry
	)

	log.Infoln("Looking for jobs in the held state...")
	if cmdOutput, err = ExecCondorQ(cl.cfg); err != nil {
		log.Errorf("%+v\n", errors.Wrap(err, "error running condor_q"))
		return
	}

	heldEntries = heldQueueEntries(cmdOutput)
	log.Infof("There are %d jobs in the held state", len(heldEntries))
	for _, entry := range heldEntries {
		if entry.InvocationID != "" {
			log.Infof("Sending stop request for invocation id %s", entry.InvocationID)
			if err = client.SendStopRequest(
				entry.InvocationID,
				"admin",
				"Job was in held state",
			); err != nil {
				log.Errorf("%+v\n", errors.Wrap(err, "error sending stop request"))
			}
		}
	}
}

func (cl *CondorLauncher) stopHandler(client *messaging.Client) func(d amqp.Delivery) {
	return func(d amqp.Delivery) {
		var (
			condorQOutput  []byte
			condorRMOutput []byte
			invID          string
			err            error
		)

		d.Ack(false)

		log.Infoln("in stopHandler")

		stopRequest := &messaging.StopRequest{}
		if err = json.Unmarshal(d.Body, stopRequest); err != nil {
			log.Errorf("%+v\n", errors.Wrap(err, "failed to unmarshal the stop request body"))
			return
		}

		invID = stopRequest.InvocationID

		log.Infoln("Running condor_q...")
		if condorQOutput, err = ExecCondorQ(cl.cfg); err != nil {
			log.Errorf("%+v\n", errors.Wrap(err, "failed to exec condor_q"))
			return
		}
		log.Infoln("Done running condor_q")

		entries := queueEntriesByInvocationID(condorQOutput, invID)
		log.Infof("Number of entries for job %s is %d", invID, len(entries))
		for _, entry := range entries {
			if entry.CondorID == "" {
				continue
			}
			condorID := entry.CondorID
			log.Infof("Running 'condor_rm %s'", condorID)
			if condorRMOutput, err = ExecCondorRm(condorID, cl.cfg); err != nil {
				log.Errorf("%+v\n", errors.Wrapf(err, "failed to run 'condor_rm %s'", condorID))
				continue
			}
			fauxJob := model.New(cl.cfg)
			fauxJob.InvocationID = invID
			update := &messaging.UpdateMessage{
				Job:     fauxJob,
				State:   messaging.FailedState,
				Message: "Job was killed",
			}
			if err = client.PublishJobUpdate(update); err != nil {
				log.Errorf("%+v\n", errors.Wrap(err, "failed to publish job update for a failed job"))
			}
			log.Infof("Output of 'condor_rm %s':\n%s", condorID, condorRMOutput)
		}
	}
}

// RegisterStopHandler registers a handler for all stop requests.
func (cl *CondorLauncher) RegisterStopHandler(client *messaging.Client) {
	exchangeName := cl.cfg.GetString("amqp.exchange.name")
	exchangeType := cl.cfg.GetString("amqp.exchange.type")
	client.AddConsumer(exchangeName, exchangeType, "condor-launcher-stops", messaging.StopRequestKey("*"), cl.stopHandler(client))
}

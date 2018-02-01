package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
)

// ExecCondorQ runs the condor_q -long command and returns its output.
func ExecCondorQ(condorPath, condorConfig string) ([]byte, error) {
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
	cmd.Env = []string{
		fmt.Sprintf("PATH=%s", condorPath),
		fmt.Sprintf("CONDOR_CONFIG=%s", condorConfig),
	}
	output, err = cmd.CombinedOutput()
	if err != nil {
		return output, errors.Wrapf(err, "failed to get the output of the command '%s %s'", csPath, "-long")
	}
	return output, nil
}

// ExecCondorRm runs condor_rm with an IpcUuid constraint for the given invocationID.
// Returns the output of the command and possibly an error.
func ExecCondorRm(invocationID, condorPath, condorConfig string) ([]byte, error) {
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

	// condor_rm -constraint 'IpcUuid =?= "<uuid>"'
	constraintIpcUuid := fmt.Sprintf(`IpcUuid =?= "%s"`, invocationID)
	cmd := exec.Command(crPath, "-constraint", constraintIpcUuid)
	cmd.Env = []string{
		fmt.Sprintf("PATH=%s", condorPath),
		fmt.Sprintf("CONDOR_CONFIG=%s", condorConfig),
	}
	output, err = cmd.CombinedOutput()
	if err != nil {
		return output, errors.Wrapf(err, "failed to get the output of '%s %s %s'", crPath, "-constraint", constraintIpcUuid)
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

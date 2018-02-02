package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// ExecCondorQHeldIDs runs the
// `condor_q -constraint 'JobStatus =?= 5' -format "%s\n" IpcUuid`
// command and returns its output.
func ExecCondorQHeldIDs(condorPath, condorConfig string) ([]byte, error) {
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

	cmdArgs := []string{
		"-constraint",
		"JobStatus =?= 5",
		"-format",
		"%s\\n",
		"IpcUuid"}

	cmd := exec.Command(csPath, cmdArgs...)
	cmd.Env = []string{
		fmt.Sprintf("PATH=%s", condorPath),
		fmt.Sprintf("CONDOR_CONFIG=%s", condorConfig),
	}
	output, err = cmd.CombinedOutput()
	if err != nil {
		return output, errors.Wrapf(err,
			"failed to get the output of the command '%s %s'",
			csPath,
			strings.Join(cmdArgs, " "))
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

func heldQueueInvocationIDs(condorQFormattedOutput []byte) []string {
	var retval []string

	lines := bytes.Split(condorQFormattedOutput, []byte("\n"))
	for _, line := range lines {
		invocationID := bytes.TrimSpace(line)
		if len(invocationID) > 0 {
			retval = append(retval, string(invocationID))
		}
	}

	return retval
}

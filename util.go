package main

import (
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

// toAbsolutePath converts a relative path to an absolute path, logging a fatal error if the path can't be converted.
func toAbsolutePath(relPath string) string {
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "failed to get the absolute path to %s", relPath))
	}
	return absPath
}

// findExecPath finds the absolute path of an executable file somewhere in the search path, logging a fatal error if
// the executable file can't be found.
func findExecPath(execName string) string {
	execPath, err := exec.LookPath(execName)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "failed to find %s in $PATH", execName))
	}
	return toAbsolutePath(execPath)
}

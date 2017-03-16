package main

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/cyverse-de/model"
	"github.com/pkg/errors"
)

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

// CreateSubmissionDirectory creates a directory for a submission and returns the path to it as a string.
func CreateSubmissionDirectory(s *model.Job) (string, error) {
	dirPath := s.CondorLogDirectory()
	if path.Base(dirPath) != "logs" {
		dirPath = path.Join(dirPath, "logs")
	}
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create the directory %s", dirPath)
	}
	return dirPath, err
}

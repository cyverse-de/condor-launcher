package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/cyverse-de/model"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
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

// CreateSubmissionFiles creates the iplant.cmd file inside the
// directory designated by 'dir'. The return values are the path to the iplant.cmd
// file, and any errors, in that order.
func CreateSubmissionFiles(dir string, cfg *viper.Viper, s *model.Job) (string, string, string, error) {
	cmdContents, err := GenerateCondorSubmit(SubmissionTemplate, s)
	if err != nil {
		return "", "", "", err
	}
	jobConfigContents, err := GenerateJobConfig(JobConfigTemplate, cfg)
	if err != nil {
		return "", "", "", err
	}
	jobContents, err := json.Marshal(s)
	if err != nil {
		return "", "", "", err
	}
	irodsContents, err := GenerateIRODSConfig(IRODSConfigTemplate, cfg)
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
			return "", "", "", errors.Wrapf(err, "failed to write to file %s", sf.filename)
		}
	}
	return subfiles[0].filename, subfiles[1].filename, subfiles[2].filename, nil
}

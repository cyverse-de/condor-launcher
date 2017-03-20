package main

import (
	"io/ioutil"
	"os"
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

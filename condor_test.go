package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"

	"gopkg.in/cyverse-de/model.v4"

	"github.com/cyverse-de/condor-launcher/test"
)

type filerecord struct {
	path     string
	contents []byte
	mode     os.FileMode
}

type tsys struct {
	dirsCreated  map[string]os.FileMode
	filesWritten []filerecord
}

// newTSys creates a new instance of tsys.
func newtsys() *tsys {
	return &tsys{
		dirsCreated:  make(map[string]os.FileMode, 0),
		filesWritten: make([]filerecord, 0),
	}
}

func (t *tsys) MkdirAll(path string, mode os.FileMode) error {
	t.dirsCreated[path] = mode
	return nil
}

func (t *tsys) WriteFile(path string, contents []byte, mode os.FileMode) error {
	t.filesWritten = append(t.filesWritten, filerecord{
		path:     path,
		contents: contents,
		mode:     mode,
	})
	return nil
}

func TestLaunch(t *testing.T) {
	cfg := test.InitConfig(t)
	test.InitPath(t)
	csPath, err := exec.LookPath("condor_submit")
	if err != nil {
		t.Error(errors.Wrapf(err, "failed to find condor_submit in $PATH"))
	}
	if !path.IsAbs(csPath) {
		csPath, err = filepath.Abs(csPath)
		if err != nil {
			t.Error(errors.Wrapf(err, "failed to get the absolute path to %s", csPath))
		}
	}
	crPath, err := exec.LookPath("condor_rm")
	if err != nil {
		t.Error(errors.Wrap(err, "failed to find condor_rm on the $PATH"))
	}
	if !path.IsAbs(crPath) {
		crPath, err = filepath.Abs(crPath)
		if err != nil {
			t.Error(errors.Wrapf(err, "failed to get the absolute path for %s", crPath))
		}
	}
	filesystem := newtsys()
	cl := New(cfg, nil, filesystem, csPath, crPath)
	if err != nil {
		t.Error(err)
	}
	data, err := ioutil.ReadFile("test/test_submission.json")
	if err != nil {
		t.Error(err)
	}
	j, err := model.NewFromData(cl.cfg, data)
	if err != nil {
		t.Error(err)
	}
	actual, err := cl.launch(j, "", "")
	if err != nil {
		t.Error(err)
	}
	expected := "10000"
	if actual != expected {
		t.Errorf("launch returned:\n%s\ninstead of:\n%s\n", actual, expected)
	}
	parent := path.Join(j.CondorLogPath, "test_this_is_a_test")
	err = os.RemoveAll(parent)
	if err != nil {
		t.Error(err)
	}
}

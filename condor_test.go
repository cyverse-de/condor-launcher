package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/cyverse-de/configurate"
	"github.com/cyverse-de/messaging"
	"github.com/cyverse-de/model"
	"github.com/streadway/amqp"

	"github.com/spf13/viper"
)

func JSONData(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	c, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return c, err
}

var (
	s   *model.Job
	cfg *viper.Viper
)

func inittestsFile(t *testing.T, filename string) *model.Job {
	var err error
	cfg, err = configurate.InitDefaults("test/test_config.yaml", configurate.JobServicesDefaults)
	if err != nil {
		t.Error(err)
	}
	cfg.Set("irods.base", "/path/to/irodsbase")
	cfg.Set("irods.host", "hostname")
	cfg.Set("irods.port", "1247")
	cfg.Set("irods.user", "user")
	cfg.Set("irods.pass", "pass")
	cfg.Set("irods.zone", "test")
	cfg.Set("irods.resc", "")
	cfg.Set("condor.log_path", "../test/tmp")
	cfg.Set("condor.porklock_tag", "test")
	cfg.Set("condor.filter_files", "foo,bar,baz,blippy")
	cfg.Set("condor.request_disk", "0")
	cfg.Set("condor.path_env_var", "/path/to/path")
	cfg.Set("condor.condor_config", "/condor/config")
	data, err := JSONData(filename)
	if err != nil {
		t.Error(err)
	}
	s, err = model.NewFromData(cfg, data)
	if err != nil {
		t.Error(err)
	}
	PATH := fmt.Sprintf("test/:%s", os.Getenv("PATH"))
	err = os.Setenv("PATH", PATH)
	if err != nil {
		t.Error(err)
	}
	return s
}

func _inittests(t *testing.T, memoize bool) *model.Job {
	if s == nil || !memoize {
		s = inittestsFile(t, "test/test_submission.json")
	}
	return s
}

func inittests(t *testing.T) *model.Job {
	return _inittests(t, true)
}

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

func TestGenerateCondorSubmit(t *testing.T) {
	s := inittests(t)

	actual, err := GenerateCondorSubmit(SubmissionTemplate, s)
	if err != nil {
		t.Error(err)
	}
	expected := `universe = vanilla
executable = /usr/local/bin/road-runner
rank = mips
requirements = (HAS_HOST_MOUNTS == True)
arguments = --config config --job job
output = script-output.log
error = script-error.log
log = condor.log
request_disk = 0
+IpcUuid = "07b04ce2-7757-4b21-9e15-0b4c2f44be26"
+IpcJobId = "generated_script"
+IpcUsername = "test_this_is_a_test"
+IpcUserGroups = {"groups:foo","groups:bar","groups:baz"}
concurrency_limits = _00000000000000000000000000000000
+IpcExe = "wc_wrapper.sh"
+IpcExePath = "/usr/local3/bin/wc_tool-1.00"
should_transfer_files = YES
transfer_input_files = irods-config,iplant.cmd,config,job
transfer_output_files = workingvolume/logs/logs-stdout-output,workingvolume/logs/logs-stderr-output
when_to_transfer_output = ON_EXIT_OR_EVICT
notification = NEVER
queue
`
	if actual.String() != expected {
		t.Errorf("GenerateCondorSubmit() returned:\n\n%s\n\ninstead of:\n\n%s", actual, expected)
	}
}

func TestGenerateCondorSubmitGroup(t *testing.T) {
	s := inittests(t)
	s.Group = "foo"
	actual, err := GenerateCondorSubmit(SubmissionTemplate, s)
	if err != nil {
		t.Error(err)
	}
	expected := `universe = vanilla
executable = /usr/local/bin/road-runner
rank = mips
requirements = (HAS_HOST_MOUNTS == True)
arguments = --config config --job job
output = script-output.log
error = script-error.log
log = condor.log
accounting_group = foo
accounting_group_user = test_this_is_a_test
request_disk = 0
+IpcUuid = "07b04ce2-7757-4b21-9e15-0b4c2f44be26"
+IpcJobId = "generated_script"
+IpcUsername = "test_this_is_a_test"
+IpcUserGroups = {"groups:foo","groups:bar","groups:baz"}
concurrency_limits = _00000000000000000000000000000000
+IpcExe = "wc_wrapper.sh"
+IpcExePath = "/usr/local3/bin/wc_tool-1.00"
should_transfer_files = YES
transfer_input_files = irods-config,iplant.cmd,config,job
transfer_output_files = workingvolume/logs/logs-stdout-output,workingvolume/logs/logs-stderr-output
when_to_transfer_output = ON_EXIT_OR_EVICT
notification = NEVER
queue
`
	if actual.String() != expected {
		t.Errorf("GenerateCondorSubmit() returned:\n\n%s\n\ninstead of:\n\n%s", actual, expected)
	}
	_inittests(t, false)
}

func TestGenerateCondorSubmitNoVolumes(t *testing.T) {
	s := inittestsFile(t, "test/no_volumes_submission.json")
	actual, err := GenerateCondorSubmit(SubmissionTemplate, s)
	if err != nil {
		t.Error(err)
	}
	expected := `universe = vanilla
executable = /usr/local/bin/road-runner
rank = mips
arguments = --config config --job job
output = script-output.log
error = script-error.log
log = condor.log
request_disk = 0
+IpcUuid = "07b04ce2-7757-4b21-9e15-0b4c2f44be26"
+IpcJobId = "generated_script"
+IpcUsername = "test_this_is_a_test"
+IpcUserGroups = {"groups:foo","groups:bar","groups:baz"}
concurrency_limits = _00000000000000000000000000000000
+IpcExe = "wc_wrapper.sh"
+IpcExePath = "/usr/local3/bin/wc_tool-1.00"
should_transfer_files = YES
transfer_input_files = irods-config,iplant.cmd,config,job
transfer_output_files = workingvolume/logs/logs-stdout-output,workingvolume/logs/logs-stderr-output
when_to_transfer_output = ON_EXIT_OR_EVICT
notification = NEVER
queue
`
	if actual.String() != expected {
		t.Errorf("GenerateCondorSubmit() returned:\n\n%s\n\ninstead of:\n\n%s", actual, expected)
	}
}

func TestCreateSubmissionDirectory(t *testing.T) {
	s := inittests(t)
	filesystem := newtsys()
	cl := New(cfg, nil, filesystem)
	dir, err := CreateSubmissionDirectory(s)
	if err != nil {
		t.Error(err)
	}
	_, err = os.Stat(dir)
	if err != nil {
		t.Error(err)
	}
	logPath := cl.cfg.GetString("condor.log_path")
	parent := path.Join(logPath, s.Submitter)
	err = os.RemoveAll(parent)
	if err != nil {
		t.Error(err)
	}
	_inittests(t, false)
}

func TestCreateSubmissionFiles(t *testing.T) {
	s := inittests(t)
	filesystem := newtsys()
	cl := New(cfg, nil, filesystem)
	dir, err := CreateSubmissionDirectory(s)
	if err != nil {
		t.Fatal(err)
	}
	cmd, sh, c, err := cl.CreateSubmissionFiles(dir, s)
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Stat(cmd)
	if err != nil {
		t.Error(err)
	}
	_, err = os.Stat(sh)
	if err != nil {
		t.Error(err)
	}
	_, err = os.Stat(c)
	if err != nil {
		t.Error(err)
	}
	irodsPath := path.Join(path.Dir(cmd), "irods-config")
	_, err = os.Stat(irodsPath)
	if err != nil {
		t.Error(err)
	}
	logPath := cl.cfg.GetString("condor.log_path")
	parent := path.Join(logPath, s.Submitter)
	err = os.RemoveAll(parent)
	if err != nil {
		t.Error(err)
	}
	_inittests(t, false)
}

func TestCondorSubmit(t *testing.T) {
	s := inittests(t)
	filesystem := newtsys()
	PATH := fmt.Sprintf(".:%s", os.Getenv("PATH"))
	err := os.Setenv("PATH", PATH)
	if err != nil {
		t.Error(err)
	}
	cl := New(cfg, nil, filesystem)
	dir, err := CreateSubmissionDirectory(s)
	if err != nil {
		t.Error(err)
	}
	cmd, _, _, err := cl.CreateSubmissionFiles(dir, s)
	if err != nil {
		t.Error(err)
	}
	actual, err := cl.submit(cmd, s)
	if err != nil {
		t.Error(err)
	}
	expected := "10000"
	if actual != expected {
		t.Errorf("CondorSubmit() returned %s instead of %s", actual, expected)
	}
	logPath := cl.cfg.GetString("condor.log_path")
	parent := path.Join(logPath, s.Submitter)
	err = os.RemoveAll(parent)
	if err != nil {
		t.Error(err)
	}
}

func TestLaunch(t *testing.T) {
	inittests(t)
	filesystem := newtsys()
	cl := New(cfg, nil, filesystem)
	data, err := JSONData("test/test_submission.json")
	if err != nil {
		t.Error(err)
	}
	j, err := model.NewFromData(cl.cfg, data)
	if err != nil {
		t.Error(err)
	}
	actual, err := cl.launch(j)
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

func TestStop(t *testing.T) {
	inittests(t)
	filesystem := newtsys()
	cl := New(cfg, nil, filesystem)
	//Start up a fake jex-events
	jr := &model.Job{
		CondorID:     "10000",
		Submitter:    "test_this_is_a_test",
		AppID:        "c7f05682-23c8-4182-b9a2-e09650a5f49b",
		InvocationID: "00000000-0000-0000-0000-000000000000",
	}
	actual, err := cl.stop(jr)
	if err != nil {
		t.Error(err)
	}
	if actual == "" {
		t.Errorf("stop returned an empty string")
	}
}

type MockConsumer struct {
	exchange     string
	exchangeType string
	queue        string
	key          string
	handler      messaging.MessageHandler
}

type MockMessage struct {
	key string
	msg []byte
}

type MockMessenger struct {
	consumers         []MockConsumer
	publishedMessages []MockMessage
	publishTo         []string
	publishError      bool
}

func (m *MockMessenger) Close()  {}
func (m *MockMessenger) Listen() {}

func (m *MockMessenger) AddConsumer(exchange, exchangeType, queue, key string, handler messaging.MessageHandler) {
	m.consumers = append(m.consumers, MockConsumer{
		exchange:     exchange,
		exchangeType: exchangeType,
		queue:        queue,
		key:          key,
		handler:      handler,
	})
}

func (m *MockMessenger) Publish(key string, msg []byte) error {
	if m.publishError {
		return errors.New("publish error")
	}
	m.publishedMessages = append(m.publishedMessages, MockMessage{key: key, msg: msg})
	return nil
}

func (m *MockMessenger) PublishJobUpdate(update *messaging.UpdateMessage) error {
	if m.publishError {
		return errors.New("publish error")
	}
	m.publishedMessages = append(m.publishedMessages, MockMessage{key: "foo", msg: []byte(update.Message)})
	return nil
}

func (m *MockMessenger) SetupPublishing(exchange string) error {

	m.publishTo = append(m.publishTo, exchange)
	return nil
}

func TestHandlePing(t *testing.T) {
	inittests(t)
	client := &MockMessenger{
		publishedMessages: make([]MockMessage, 0),
	}
	filesystem := newtsys()
	launcher := New(cfg, client, filesystem)
	delivery := amqp.Delivery{
		RoutingKey: "events.condor-launcher.ping",
	}
	launcher.handlePing(delivery)
	mm := launcher.client.(*MockMessenger)
	if len(mm.publishedMessages) != 1 {
		t.Errorf("number of published messages was %d instead of 1", len(mm.publishedMessages))
	}
	if mm.publishedMessages[0].key != pongKey {
		t.Errorf("routing key was %s instead of %s", mm.publishedMessages[0].key, pongKey)
	}
}

func TestHandleEvents(t *testing.T) {
	inittests(t)
	client := &MockMessenger{
		publishedMessages: make([]MockMessage, 0),
	}
	filesystem := newtsys()
	launcher := New(cfg, client, filesystem)
	delivery := amqp.Delivery{
		RoutingKey: "events.condor-launcher.ping",
	}
	launcher.handleEvents(delivery)
	mm := launcher.client.(*MockMessenger)
	if len(mm.publishedMessages) != 1 {
		t.Errorf("number of published messages was %d instead of 1", len(mm.publishedMessages))
	}
	if mm.publishedMessages[0].key != pongKey {
		t.Errorf("routing key was %s instead of %s", mm.publishedMessages[0].key, pongKey)
	}
}

func TestHandleBadEvents(t *testing.T) {
	inittests(t)
	client := &MockMessenger{
		publishedMessages: make([]MockMessage, 0),
	}
	filesystem := newtsys()
	launcher := New(cfg, client, filesystem)
	delivery := amqp.Delivery{
		RoutingKey: "events.condor-launcher.pinadsfasdfg",
	}
	launcher.handleEvents(delivery)
	mm := launcher.client.(*MockMessenger)
	if len(mm.publishedMessages) != 0 {
		t.Errorf("number of published messages was %d instead of 1", len(mm.publishedMessages))
	}
}

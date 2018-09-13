package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"

	"github.com/streadway/amqp"
	"gopkg.in/cyverse-de/messaging.v6"
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

type VaultTester struct{}

func (v *VaultTester) MountCubbyhole(mountPoint string) error {
	return nil
}

func (v *VaultTester) ChildToken(numUses int) (string, error) {
	return "", nil
}

func (v *VaultTester) StoreConfig(token, mountPoint, jobID string, config []byte) error {
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
	cl.v = &VaultTester{}
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

func (m *MockMessenger) Close()                        {}
func (m *MockMessenger) Listen()                       {}
func (m *MockMessenger) DeleteQueue(name string) error { return nil }

func (m *MockMessenger) AddConsumer(exchange, exchangeType, queue, key string, handler messaging.MessageHandler, prefetchCount int) {
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

func TestHandleEvents(t *testing.T) {
	cfg := test.InitConfig(t)
	test.InitPath(t)
	client := &MockMessenger{
		publishedMessages: make([]MockMessage, 0),
	}
	filesystem := newtsys()
	launcher := New(cfg, client, filesystem, "condor_submit", "condor_rm")
	delivery := amqp.Delivery{
		RoutingKey: "events.condor-launcher.ping",
	}
	launcher.routeEvents(delivery)
	mm := launcher.client.(*MockMessenger)
	if len(mm.publishedMessages) != 1 {
		t.Errorf("number of published messages was %d instead of 1", len(mm.publishedMessages))
	}
	if mm.publishedMessages[0].key != pongKey {
		t.Errorf("routing key was %s instead of %s", mm.publishedMessages[0].key, pongKey)
	}
}

func TestHandleBadEvents(t *testing.T) {
	cfg := test.InitConfig(t)
	test.InitPath(t)
	client := &MockMessenger{
		publishedMessages: make([]MockMessage, 0),
	}
	filesystem := newtsys()
	launcher := New(cfg, client, filesystem, "condor_submit", "condor_rm")
	delivery := amqp.Delivery{
		RoutingKey: "events.condor-launcher.pinadsfasdfg",
	}
	launcher.routeEvents(delivery)
	mm := launcher.client.(*MockMessenger)
	if len(mm.publishedMessages) != 0 {
		t.Errorf("number of published messages was %d instead of 1", len(mm.publishedMessages))
	}
}

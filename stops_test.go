package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/cyverse-de/messaging.v6"

	"github.com/cyverse-de/condor-launcher/test"
	"github.com/streadway/amqp"
)

func shouldrun() bool {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "" {
		return true
	}
	return false
}

var (
	listing = []byte(`
63c5523d-d8a5-49bc-addc-99a73566cd89
b788569f-6948-4586-b5bd-5ea096986331
eca67a7c-e745-4e98-b892-67a9948bc2cb

571722aa-f46c-40fd-a688-69068fb52ee1
105fba2c-c9a5-4a89-8033-9d3b3e5c419f
51d4cb60-e925-4229-a257-8d5b6feeedb8
22c13517-4a2d-4874-92b3-5b5d03299d60
316bc5ba-aaee-4922-a8c7-ea9479a65650
1e0cc85e-6053-452c-b1a9-cedc0f12cf7b
7ffa2752-e7bf-4eae-8ed5-b366aef7978e
d9eaa702-14d5-439e-b01f-4ca3a39096ec
57ff5e6b-5a4f-496b-9f66-eeadfbd03e86
0af6c62a-3570-46de-9f27-9b6c430bd912
0d607f70-ad65-4a8c-bfe3-1b0efb7a20e9
10af43af-6f63-4305-8bad-e70ae208ec4c
1fa4bea3-8f3e-472e-9981-3b1a5b0d425e
30282173-70e5-4055-8621-a3a149db1829
9b21c5c6-19b3-45cc-9e87-720d57032e71
6a54a3bc-1a42-453d-9fa9-81300804f26f
827624c9-1e1b-4783-ab7a-23e78c599cc3
c6193b50-cdaa-48c8-aa69-3957431b05a2
d0952a64-8d7c-4d2b-9046-6c91425e0671

4457ec2d-5203-4ab6-95e5-2831d998dd1f
6e7d9735-1fc7-4626-b5d3-91993eae7a1b
3c967e99-51ff-484d-a000-3ce3adf744ce
8d64bc74-750a-4cd3-bbf3-2d9a4e56fcf0
77ac1023-c6d1-403c-8088-04205a270c7a
dc9f2455-da10-42c6-bc58-b7a963b46338
74a17e36-ac52-4dde-9abb-e5a8dca0b2c6
7446aa7b-16db-4b4f-ad8d-8d802b402aaa
ee6cdfe9-bd90-460a-bb5f-b10b3e299b10
a45770ad-cc7d-4fc8-8b1a-ffb75bd44e0e
dd632e6c-50e7-4b81-bb07-9d2abe0a9b15
f532df1e-591e-47f1-84e6-fdf670727040
12febe03-ae83-48f2-b828-c2865fecf09e
e802a7b5-13f2-462d-ac85-f9836a4fd575
18f6b09f-63bd-4825-a34e-4767dee4f358
4a039e5a-0335-45ee-880b-5495c92bfc9a
6123e600-ec4f-4d29-b2c4-edd0af089b04
ba30efdd-4207-4120-91dd-de5caed467ae
9d20a879-6d33-43b4-b430-0c1cfaf5a7c4
8f4002c0-d650-4252-943d-56287b6c6efe
3c5c986e-1b69-4ff2-a470-0f677148240e
f0ffb7ee-b7c6-4a07-814b-88d76b3cc3b8
65ecca14-759e-41ee-b764-069ac0155df5
36770cc9-d80b-405c-b9be-c5977ae3e83b
4a4926a6-2a18-4b37-a361-581658a0f067
15384749-bc5f-43ff-854e-b6495e0c6ee4
2e8c0c9c-133a-4436-b1a4-3bb303ce7cd3`)
)

func isInvocationIDHeld(condorQListing []byte, invocationID string) bool {
	actual := heldQueueInvocationIDs(condorQListing)
	for _, uuid := range actual {
		if uuid == invocationID {
			return true
		}
	}

	return false
}

func TestCondorID(t *testing.T) {
	invID := "63c5523d-d8a5-49bc-addc-99a73566cd89"
	found := isInvocationIDHeld(listing, invID)
	if !found {
		t.Errorf("The expected InvocationID of %s was not in the Held state", invID)
	}

	invID = "b788569f-6948-4586-b5bd-5ea096986331"
	found = isInvocationIDHeld(listing, invID)
	if !found {
		t.Errorf("The expected InvocationID of %s was not in the Held state", invID)
	}

	invID = "eca67a7c-e745-4e98-b892-67a9948bc2cb"
	found = isInvocationIDHeld(listing, invID)
	if !found {
		t.Errorf("The expected InvocationID of %s was not in the Held state", invID)
	}
}

func TestExecCondorQ(t *testing.T) {
	test.InitPath(t)
	output, err := ExecCondorQHeldIDs("", "")
	if err != nil {
		t.Error(err)
	}

	invID := "63c5523d-d8a5-49bc-addc-99a73566cd89"
	found := isInvocationIDHeld(output, invID)
	if !found {
		t.Errorf("The expected InvocationID of %s was not in the Held state", invID)
	}

	invID = "b788569f-6948-4586-b5bd-5ea096986331"
	found = isInvocationIDHeld(output, invID)
	if !found {
		t.Errorf("The expected InvocationID of %s was not in the Held state", invID)
	}

	invID = "eca67a7c-e745-4e98-b892-67a9948bc2cb"
	found = isInvocationIDHeld(output, invID)
	if !found {
		t.Errorf("The expected InvocationID of %s was not in the Held state", invID)
	}
}

func TestExecCondorRm(t *testing.T) {
	test.InitPath(t)
	actual, err := ExecCondorRm("foo", "", "")
	if err != nil {
		t.Error(err)
	}
	expected := []byte("IpcUuid =?= \"foo\" was stopped\n")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("ExecCondorRm returned '%s' instead of '%s'", actual, expected)
	}
}

func TestStopHandler(t *testing.T) {
	var (
		coord      chan string
		err        error
		marshalled []byte
	)
	if !shouldrun() {
		return
	}
	cfg := test.InitConfig(t)
	test.InitPath(t)
	filesystem := newtsys()
	cl := New(cfg, nil, filesystem, "condor_submit", "condor_rm")
	stopMsg := messaging.StopRequest{
		InvocationID: "b788569f-6948-4586-b5bd-5ea096986331",
	}
	if marshalled, err = json.Marshal(stopMsg); err != nil {
		t.Error(err)
	}
	msg := amqp.Delivery{
		Body: marshalled,
	}
	old := os.Stdout
	defer func() {
		os.Stdout = old
	}()
	r, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stdout = w
	coord = make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		coord <- buf.String()
	}()
	cl.stopHandler("", "")(msg)
	w.Close()
	actual := <-coord
	if !strings.Contains(actual, "Running condor_q...") {
		t.Error("Logging output from stopHandler does not contain 'Running condor_q...'")
	}
	if !strings.Contains(actual, "Done running condor_q") {
		t.Error("Logging output from stopHandler does not contain 'Done running condor_q'")
	}
	if !strings.Contains(actual, "Running 'condor_rm 1'") {
		t.Error("Logging output from stopHandler does not contain \"Running 'condor_rm 1'\"")
	}
	if !strings.Contains(actual, "Output of 'condor_rm 1'") {
		t.Error("Logging output from stopHandler does not contain \"Output of 'condor_rm 1'\"")
	}
}

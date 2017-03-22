package main

import "testing"

func TestVaultInit(t *testing.T) {
	v, err := VaultInit("test", "https://test.test1.test2:8200")
	if err != nil {
		t.Error(err)
	}
	if v == nil {
		t.Error("client was nil")
	}
	cfg := v.api.GetConfig()
	if cfg.Address != "https://test.test1.test2:8200" {
		t.Errorf("address was %s instead of 'test'", cfg.Address)
	}
	actual := v.api.Client().Token()
	expected := "test"
	if actual != expected {
		t.Errorf("token was '%s' instead of '%s'", actual, expected)
	}

	v, err = VaultInit("test", "asdfasdf")
	if err == nil {
		t.Error("No url parse error")
	}
	if v != nil {
		t.Error("client was not nil")
	}
}

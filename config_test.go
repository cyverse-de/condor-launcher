package main

import (
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

func TestCopyConfig(t *testing.T) {
	v1 := viper.New()
	v1.Set("test0", "value0")
	v1.Set("test1", "value1")
	v1.Set("test2", "value2")
	v2 := CopyConfig(v1)
	actual := v2.AllSettings()
	expected := v1.AllSettings()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("copied settings %#v not equal to original settings %#v", actual, expected)
	}
	v1.Set("test1", "newvalue")
	expected1 := "value1"
	actual1 := v2.Get("test1")

	if actual1 != expected1 {
		t.Errorf("value was %s instead of %s", actual, expected)
	}

}

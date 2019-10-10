package submitfile

import (
	"testing"
)

func checkList(t *testing.T, l []string, expected string) {
	actual := FormatList(l)
	if actual != expected {
		t.Errorf("Unexpected list format: actual `%s`; expected `%s`", actual, expected)
	}
}

func TestFormatList(t *testing.T) {
	checkList(t, []string{}, `{}`)
	checkList(t, []string{"\t\n\f\r \"'\\"}, `{"\t\n\f\r \"\'\\"}`)
	checkList(t, []string{"foo", "bar", "baz"}, `{"foo","bar","baz"}`)
	checkList(t, []string{"foo\t", "bar"}, `{"foo\t","bar"}`)
	checkList(t, []string{"foo\n", "bar"}, `{"foo\n","bar"}`)
	checkList(t, []string{"foo\f", "bar"}, `{"foo\f","bar"}`)
	checkList(t, []string{"foo\r", "bar"}, `{"foo\r","bar"}`)
	checkList(t, []string{"foo\"", "bar"}, `{"foo\"","bar"}`)
	checkList(t, []string{"foo'", "bar"}, `{"foo\'","bar"}`)
	checkList(t, []string{"foo\\", "bar"}, `{"foo\\","bar"}`)
	checkList(t, []string{"foo", "bar\t"}, `{"foo","bar\t"}`)
	checkList(t, []string{"foo", "bar\n"}, `{"foo","bar\n"}`)
	checkList(t, []string{"foo", "bar\f"}, `{"foo","bar\f"}`)
	checkList(t, []string{"foo", "bar\r"}, `{"foo","bar\r"}`)
	checkList(t, []string{"foo", "bar\""}, `{"foo","bar\""}`)
	checkList(t, []string{"foo", "bar'"}, `{"foo","bar\'"}`)
	checkList(t, []string{"foo", "bar\\"}, `{"foo","bar\\"}`)
	checkList(t, []string{"groups:foo", "groups:bar"}, `{"groups:foo","groups:bar"}`)
}

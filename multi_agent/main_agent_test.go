package multi_agent

import (
	"reflect"
	"testing"
)

func TestNormalizeBase64List(t *testing.T) {
	input := []string{"  abc  ", "", "   ", "\tdef\n", "ghi"}
	got := normalizeBase64List(input)
	want := []string{"abc", "def", "ghi"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected normalize result: want=%v got=%v", want, got)
	}
}


package types

import (
	"bytes"
	"testing"
)

// TestYAMLEncoder verifies YAMLEncoder's own behavior: it rewrites builder.go's
// default markers ("zzz#(desc)(type)name" keys) into YAML comments, and it
// propagates marshal errors.
func TestYAMLEncoder(t *testing.T) {
	var buf bytes.Buffer
	if err := YAMLEncoder(&buf, map[string]string{"zzz#(a description)(string)fieldName": "somevalue"}); err != nil {
		t.Fatalf("YAMLEncoder() error = %v", err)
	}
	if got, want := buf.String(), "# fieldName: somevalue\n"; got != want {
		t.Errorf("YAMLEncoder() = %q, want %q", got, want)
	}

	if err := YAMLEncoder(&bytes.Buffer{}, func() {}); err == nil {
		t.Error("YAMLEncoder() expected error for unmarshalable value, got nil")
	}
}

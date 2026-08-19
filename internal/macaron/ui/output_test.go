package ui

import (
	"bytes"
	"io"
	"testing"
)

func TestServiceOutputIsDimmedOnTerminal(t *testing.T) {
	var out bytes.Buffer
	output := &Output{out: &out, err: io.Discard, useANSI: true}
	output.Service("api", "ready")
	if got := out.String(); got != "\x1b[2m🪵 api  ready\x1b[0m\n" {
		t.Fatalf("service output = %q", got)
	}
}

package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseArgsRequiresRange(t *testing.T) {
	t.Parallel()

	_, err := parseArgs([]string{"--output", "x.html"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for missing positional args")
	}
	if !strings.Contains(err.Error(), "<from-ref>") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestParseArgsVersion(t *testing.T) {
	t.Parallel()

	cfg, err := parseArgs([]string{"--version"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("parseArgs() error = %v", err)
	}
	if !cfg.showVer {
		t.Fatalf("expected showVer=true")
	}
}

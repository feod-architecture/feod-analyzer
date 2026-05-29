package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeCommandWritesJSON(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "valid")
	out := filepath.Join(t.TempDir(), "report")

	code := run([]string{"analyze", root, "--out", out, "--formats", "json", "--fail-on", "error"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	if _, err := os.Stat(filepath.Join(out, "feod-report.json")); err != nil {
		t.Fatalf("expected JSON report: %v", err)
	}
}

func TestAnalyzeCommandFailOnError(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "violations")
	out := filepath.Join(t.TempDir(), "report")

	code := run([]string{"analyze", root, "--out", out, "--formats", "json", "--fail-on", "error"})
	if code != 1 {
		t.Fatalf("expected exit code 1 for violations, got %d", code)
	}
}

func TestAnalyzeCommandRejectsUnknownFormat(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "valid")
	out := filepath.Join(t.TempDir(), "report")

	code := run([]string{"analyze", root, "--out", out, "--formats", "xml", "--fail-on", "error"})
	if code != 2 {
		t.Fatalf("expected exit code 2 for invalid format, got %d", code)
	}
}

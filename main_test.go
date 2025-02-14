package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestParseWorkflowFile(t *testing.T) {
	// Create a test YAML file
	yamlContent := `name: Test Workflow
jobs:
  build:
    name: Build Job
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
  test:
    runs-on: ubuntu-latest
    steps:
      - run: go test ./...`

	tmpfile, err := os.CreateTemp("", "workflow*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(yamlContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Set command line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", tmpfile.Name()}

	// Capture output
	output := captureOutput(func() {
		err := run()
		if err != nil {
			t.Fatal(err)
		}
	})

	// Split output into lines and remove empty lines
	lines := []string{}
	for _, line := range strings.Split(output, "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}

	// Expected output lines
	expectedLines := []string{
		"File: " + tmpfile.Name(),
		"Workflow: Test Workflow",
		"Job ID: build",
		"Job Name: Build Job",
		"Job ID: test",
	}

	if len(lines) != len(expectedLines) {
		t.Errorf("expected %d lines, got %d lines\nexpected:\n%v\ngot:\n%v",
			len(expectedLines), len(lines), expectedLines, lines)
		return
	}

	for i, expected := range expectedLines {
		if lines[i] != expected {
			t.Errorf("line %d:\nexpected: %q\ngot: %q", i+1, expected, lines[i])
		}
	}
}

func TestParseWorkflowDirectory(t *testing.T) {
	// Create test directory structure
	tmpDir, err := os.MkdirTemp("", "workflow-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .github/workflows directory
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create two workflow files
	workflow1 := `name: Workflow 1
jobs:
  job1:
    name: Job 1
    runs-on: ubuntu-latest
    steps:
      - run: echo "test"`

	workflow2 := `name: Workflow 2
jobs:
  job2:
    name: Job 2
    runs-on: ubuntu-latest
    steps:
      - run: echo "test"`

	if err := os.WriteFile(filepath.Join(workflowsDir, "workflow1.yml"), []byte(workflow1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workflowsDir, "workflow2.yaml"), []byte(workflow2), 0644); err != nil {
		t.Fatal(err)
	}

	// Set command line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", tmpDir}

	// Capture output
	output := captureOutput(func() {
		err := run()
		if err != nil {
			t.Fatal(err)
		}
	})

	// Split output into lines and remove empty lines
	lines := []string{}
	for _, line := range strings.Split(output, "\n") {
		if line != "" && line != "----------------------------------------" {
			lines = append(lines, line)
		}
	}

	// Expected output lines
	expectedLines := []string{
		"File: " + filepath.Join(workflowsDir, "workflow1.yml"),
		"Workflow: Workflow 1",
		"Job ID: job1",
		"Job Name: Job 1",
		"File: " + filepath.Join(workflowsDir, "workflow2.yaml"),
		"Workflow: Workflow 2",
		"Job ID: job2",
		"Job Name: Job 2",
	}

	if len(lines) != len(expectedLines) {
		t.Errorf("expected %d lines, got %d lines\nexpected:\n%v\ngot:\n%v",
			len(expectedLines), len(lines), expectedLines, lines)
		return
	}

	for i, expected := range expectedLines {
		if lines[i] != expected {
			t.Errorf("line %d:\nexpected: %q\ngot: %q", i+1, expected, lines[i])
		}
	}
}

func TestInvalidDirectory(t *testing.T) {
	// Specify non-existent directory
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "/path/to/nonexistent"}

	if err := run(); err == nil {
		t.Error("expected error for non-existent directory, but got none")
	}
}

func TestInvalidYaml(t *testing.T) {
	// Create invalid YAML file
	yamlContent := `invalid: yaml: content
  - not properly formatted`

	tmpfile, err := os.CreateTemp("", "invalid*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(yamlContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Set command line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", tmpfile.Name()}

	// Execute run() and check for error
	if err := run(); err == nil {
		t.Error("expected error for invalid YAML file, but got none")
	}
}

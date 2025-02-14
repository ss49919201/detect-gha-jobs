package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// workflowFile represents the entire GitHub Actions workflow
type workflowFile struct {
	Name string                `yaml:"name"`
	On   interface{}           `yaml:"on"`
	Jobs map[string]*jobConfig `yaml:"jobs"`
}

// jobConfig represents the configuration of a job
type jobConfig struct {
	Name        string        `yaml:"name"`
	RunsOn      interface{}   `yaml:"runs-on"`
	Environment interface{}   `yaml:"environment,omitempty"`
	Steps       []interface{} `yaml:"steps"`
	Needs       interface{}   `yaml:"needs,omitempty"`
	If          string        `yaml:"if,omitempty"`
}

// processWorkflowFile processes a single workflow file
func processWorkflowFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	var workflow workflowFile
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return fmt.Errorf("failed to parse YAML: %v", err)
	}

	fmt.Printf("File: %s\n", filePath)
	fmt.Printf("Workflow: %s\n", workflow.Name)

	for jobID, jobConfig := range workflow.Jobs {
		fmt.Printf("Job ID: %s\n", jobID)
		if jobConfig.Name != "" {
			fmt.Printf("Job Name: %s\n", jobConfig.Name)
		}
		fmt.Println()
	}

	return nil
}

// findWorkflowFiles recursively searches for workflow files in the directory
func findWorkflowFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Target .yml or .yaml files in .github/workflows directory
		if strings.Contains(path, ".github/workflows") &&
			(strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: detect-gha-jobs <workflow-file-or-directory>")
	}

	path := os.Args[1]
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %v", err)
		}
		path = filepath.Join(cwd, path)
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to get path info: %v", err)
	}

	if info.IsDir() {
		// Process recursively if it's a directory
		files, err := findWorkflowFiles(path)
		if err != nil {
			return fmt.Errorf("failed to search workflow files: %v", err)
		}
		if len(files) == 0 {
			return fmt.Errorf("no workflow files found")
		}
		for _, file := range files {
			fmt.Println("----------------------------------------")
			if err := processWorkflowFile(file); err != nil {
				fmt.Fprintf(os.Stderr, "warning: error processing %s: %v\n", file, err)
				continue
			}
		}
	} else {
		// Process directly if it's a single file
		if err := processWorkflowFile(path); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runner/workflow.go handles parsing the GitHub Actions YAML file.
//
// A GitHub Actions workflow file looks like this:
//
//   name: CI
//   on: [push]
//   jobs:
//     build:
//       runs-on: ubuntu-latest
//       steps:
//         - name: Checkout
//           uses: actions/checkout@v3
//         - name: Run tests
//           run: go test ./...
//
// We parse it into Go structs so the rest of the code can work with it.
package runner

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Workflow is the top-level structure of a .github/workflows/*.yml file.
type Workflow struct {
	Name string         `yaml:"name"`
	Jobs map[string]Job `yaml:"jobs"`
}

// Job is a collection of steps that run on the same machine (Docker container).
// The key insight: all steps in a job share one container, so state persists
// between steps (files you create in step 1 are still there in step 2).
type Job struct {
	RunsOn string `yaml:"runs-on"`
	Steps  []Step `yaml:"steps"`
}

// Step is a single unit of work — either a shell command (`run`) or
// a pre-built action (`uses`). For v1 we support `run` steps fully
// and skip `uses` steps with a notice.
type Step struct {
	Name string            `yaml:"name"`
	Run  string            `yaml:"run"`  // shell commands to execute
	Uses string            `yaml:"uses"` // e.g. "actions/checkout@v3" — skipped in v1
	Env  map[string]string `yaml:"env"`  // environment variables for this step
}

// findWorkflow resolves which workflow file to use.
// If the user passed a path, we use that. Otherwise we auto-discover
// the first .yml file in .github/workflows/.
func findWorkflow(path string) (string, error) {
	if path != "" {
		return path, nil
	}

	matches, _ := filepath.Glob(".github/workflows/*.yml")
	if len(matches) == 0 {
		matches, _ = filepath.Glob(".github/workflows/*.yaml")
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no workflow files found in .github/workflows/ — pass a file path explicitly")
	}

	return matches[0], nil
}

// parseWorkflow reads and parses a workflow YAML file.
func parseWorkflow(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var wf Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	if len(wf.Jobs) == 0 {
		return nil, fmt.Errorf("no jobs found in %s", path)
	}

	return &wf, nil
}

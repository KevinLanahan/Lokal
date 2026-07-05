package runner

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const githubActionsTemplate = `name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      - name: Check out code
        uses: actions/checkout@v4

      - name: Install dependencies
        run: echo "Add your install command here"

      - name: Test
        run: echo "Add your test command here"

      - name: Build
        run: echo "Add your build command here"
`

const gitlabCITemplate = `image: ubuntu:22.04

stages:
  - test
  - build

test:
  stage: test
  script:
    - echo "Add your test command here"

build:
  stage: build
  script:
    - echo "Add your build command here"
`

const circleCITemplate = `version: 2.1

jobs:
  build:
    docker:
      - image: cimg/base:stable
    steps:
      - checkout
      - run:
          name: Install dependencies
          command: echo "Add your install command here"
      - run:
          name: Test
          command: echo "Add your test command here"
      - run:
          name: Build
          command: echo "Add your build command here"

workflows:
  main:
    jobs:
      - build
`

type platform struct {
	name     string
	path     string
	template string
}

var platforms = []platform{
	{name: "GitHub Actions", path: ".github/workflows/ci.yml", template: githubActionsTemplate},
	{name: "GitLab CI", path: ".gitlab-ci.yml", template: gitlabCITemplate},
	{name: "CircleCI", path: ".circleci/config.yml", template: circleCITemplate},
}

func Init() error {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println()
	fmt.Println("  lokal init — scaffold a CI workflow file")
	fmt.Println()

	// Check if any workflow already exists.
	for _, p := range platforms {
		if _, err := os.Stat(p.path); err == nil {
			fmt.Printf("  Found existing workflow: %s\n", p.path)
			fmt.Print("  Overwrite it? [y/N] > ")
			if scanner.Scan() {
				answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
				if answer != "y" && answer != "yes" {
					fmt.Println("  Cancelled.")
					return nil
				}
			}
			return writeWorkflow(p)
		}
	}

	// No existing workflow — ask which platform.
	fmt.Println("  Which CI platform are you using?")
	fmt.Println()
	for i, p := range platforms {
		fmt.Printf("    [%d] %s\n", i+1, p.name)
	}
	fmt.Println()
	fmt.Print("  Choose (1/2/3) > ")

	var choice int
	if scanner.Scan() {
		switch strings.TrimSpace(scanner.Text()) {
		case "1":
			choice = 0
		case "2":
			choice = 1
		case "3":
			choice = 2
		default:
			return fmt.Errorf("invalid choice — enter 1, 2, or 3")
		}
	}

	return writeWorkflow(platforms[choice])
}

func writeWorkflow(p platform) error {
	// Create parent directories if needed.
	dir := p.path[:strings.LastIndex(p.path, "/")]
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	if err := os.WriteFile(p.path, []byte(p.template), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", p.path, err)
	}

	fmt.Printf("\n  ✓  Created %s\n", p.path)
	fmt.Printf("  Edit it to match your project, then run:\n\n")
	fmt.Printf("    ./lokal run\n\n")
	return nil
}

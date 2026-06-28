// cmd/root.go sets up the CLI using Cobra.
// Cobra is the standard Go library for building CLI tools (used by kubectl, gh, Hugo, etc.)
// It handles argument parsing, help text, and subcommand routing for us.
package cmd

import (
	"fmt"
	"os"

	"github.com/KevinLanahan/cidb/runner"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cidb",
	Short: "cidb — step-through debugger for GitHub Actions pipelines",
	Long: `cidb lets you run a GitHub Actions workflow file locally,
pausing before each step so you can inspect, skip, retry, or
drop into a live shell inside the running container.`,
}

var runCmd = &cobra.Command{
	Use:   "run [workflow-file]",
	Short: "Run a workflow step-by-step",
	Long: `Parses the given workflow file (or auto-discovers one in .github/workflows/)
and runs each step inside a local Docker container, pausing for your input before each one.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workflowFile := ""
		if len(args) > 0 {
			workflowFile = args[0]
		}
		return runner.Run(workflowFile)
	},
}

// Execute is called by main.go to start the CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
}

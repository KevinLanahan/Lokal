package cmd

import (
	"github.com/KevinLanahan/lokal/runner"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffold a starter workflow file for your project",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runner.Init()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

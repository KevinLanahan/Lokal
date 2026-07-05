package cmd

import (
	"fmt"
	"os"

	"github.com/KevinLanahan/lokal/runner"
	"github.com/spf13/cobra"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage secrets for your local pipeline runs",
}

var secretsSetCmd = &cobra.Command{
	Use:   "set KEY=VALUE",
	Short: "Set a secret in .env",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runner.SecretsSet(args[0])
	},
}

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secrets (values masked)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runner.SecretsList()
	},
}

var secretsRemoveCmd = &cobra.Command{
	Use:   "remove KEY",
	Short: "Remove a secret from .env",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runner.SecretsRemove(args[0])
	},
}

var secretsImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import secrets from your current shell environment",
	Long:  `Scans your shell environment for common CI secret patterns and imports them into .env.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		patterns := []string{
			"AWS_", "GCP_", "GOOGLE_", "GITHUB_TOKEN", "NPM_TOKEN",
			"DOCKER_", "DATABASE_URL", "REDIS_URL", "SECRET", "API_KEY", "TOKEN",
		}
		imported := 0
		for _, env := range os.Environ() {
			for _, p := range patterns {
				if len(env) > len(p) && env[:len(p)] == p {
					if err := runner.SecretsSet(env); err == nil {
						imported++
					}
					break
				}
			}
		}
		if imported == 0 {
			fmt.Println("  No matching secrets found in current environment.")
		} else {
			fmt.Printf("  ✓  Imported %d secret(s) into .env\n", imported)
		}
		return nil
	},
}

func init() {
	secretsCmd.AddCommand(secretsSetCmd)
	secretsCmd.AddCommand(secretsListCmd)
	secretsCmd.AddCommand(secretsRemoveCmd)
	secretsCmd.AddCommand(secretsImportCmd)
	rootCmd.AddCommand(secretsCmd)
}

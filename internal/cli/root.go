package cli

import (
	"github.com/spf13/cobra"
)

var cfgPath string

var rootCmd = &cobra.Command{
	Use:   "aiguard",
	Short: "Local-first AI development harness",
	Long:  "AIGuard prevents AI-generated code from violating business requirements by gating the development workflow with deterministic checks and multi-reviewer verification.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "path to .aiguard/config.yaml (overrides auto-detection)")
}

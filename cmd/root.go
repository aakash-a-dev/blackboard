package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "blackboard",
	Short: "Spin up a mock HTTP server from a YAML file",
	Long: `blackboard — zero-boilerplate mock API server.

Write a YAML file, run blackboard serve, get a live server.
No code, no account, no runtime required.

Documentation: https://github.com/aakash-a-dev/blackboard`,
}

func Execute() {
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	root.AddCommand(serveCmd)
	root.AddCommand(validateCmd)
	root.AddCommand(initCmd)
}

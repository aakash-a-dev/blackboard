package cmd

import (
	"fmt"

	"github.com/mockapi/mockapi/internal/config"
	"github.com/mockapi/mockapi/internal/logger"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate <file>",
	Short: "Validate YAML schema and report errors, then exit",
	Args:  cobra.ExactArgs(1),
	RunE:  runValidate,
}

func runValidate(cmd *cobra.Command, args []string) error {
	path := args[0]

	_, warnings, err := config.Load(path)
	for _, w := range warnings {
		logger.Warn(w)
	}
	if err != nil {
		return err
	}

	fmt.Println("  OK — " + path + " is valid")
	return nil
}

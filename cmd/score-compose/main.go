package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/score-spec/score-compose/internal/cmd"
	"github.com/score-spec/score-compose/internal/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	VerbosityLevelHTTPRequest      = 8
	VerbosityLevelHTTPRequestDebug = 9
)

func main() {
	var (
		configFilePath string
		c              config.Config
		rootCmd        *cobra.Command
	)

	scoreCmds := cmd.New(&c, afero.NewOsFs())

	rootCmd = &cobra.Command{
		Use:   cmd.CLIName + " [verb]",
		Short: cmd.CLIName + " is an implementation of Score for the Compose Spec",
		Long: cmd.CLIName + ` is an implementation of Score for Compose,

Find more information at https://score.dev
`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "help" || cmd.Name() == "version" {
				return nil
			}
			// Don't show usage whenever an error is returned from a command.
			// NOTE: Using the help command and --help still works
			// As described: https://github.com/spf13/cobra/issues/340#issuecomment-378726225
			cmd.SilenceUsage = true

			// This function is called once per command once the persistent flags have been parsed.
			// It allows us to sync up the configuration which is managed with viper with the flags
			// which are managed with cobra.
			// From https://pkg.go.dev/github.com/spf13/cobra#Command:
			//   children of this command will inherit and execute.

			if err := config.Load(&c, rootCmd.PersistentFlags()); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("unable to load specified config file: %s", configFilePath)
				}
				return err
			}

			return nil
		},
	}

	// Persistent Flags are available to all commands. We treat them as configuration overrides.
	rootCmd.PersistentFlags().String("context-dir", "", "Overrides the .score-compose directory used for context.")

	rootCmd.AddCommand(cmd.NewCmdInit(scoreCmds))
	rootCmd.AddCommand(cmd.NewCmdGenerate(scoreCmds))

	usedCmd, err := rootCmd.ExecuteC()
	if err != nil {
		if errors.Is(err, cmd.ErrUsage) {
			usedCmd.Usage()
		}
		os.Exit(1)
	}
}

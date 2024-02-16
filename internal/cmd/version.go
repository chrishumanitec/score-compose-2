package cmd

import (
	"fmt"

	"github.com/score-spec/score-compose/internal/version"
	"github.com/spf13/cobra"
)

func NewCmdVersion(c Commands) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display version of " + CLIName,
		Long:  "Display version of " + CLIName,
		Example: `# Output the current version
` + CLIName + ` version`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s version %s\n", CLIName, version.Version)
		},
	}
}

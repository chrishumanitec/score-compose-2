package cmd

import (
	"github.com/spf13/cobra"
)

func NewCmdInit(c Commands) *cobra.Command {

	initCmd := &cobra.Command{
		Use:   "init PROJECTNAME",
		Short: "Creates a score compose project.",
		Long: `Creates a new score compose project.

PROJECTNAME  is the name that will be given to the compose project.
             It must be lowercase alphanumeric with - and _. It cannot start
			with a - or _.
             See: https://github.com/compose-spec/compose-spec/blob/master/04-version-and-name.md#name-top-level-element
`,
		Example: `# Create a new score compose project in the current diretory
` + CLIName + ` init my-score-compose-project

# Create a new score compose project in the parent directory
` + CLIName + ` init my-score-compose-project --current-dir ../`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Init(args[0])
		},
		Args: cobra.ExactArgs(1),
	}
	return initCmd
}

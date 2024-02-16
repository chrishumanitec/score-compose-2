package cmd

import (
	"github.com/spf13/cobra"
)

func NewCmdGenerate(c Commands) *cobra.Command {
	var (
		propertyOverrides []string
		overridesFile     string
		buildStr          string
		scoreFilePath     string
	)
	generateCmd := &cobra.Command{
		Use:   "generate [(-p|--property)PROPERTYPATH[=PROPERTYVALUE]]",
		Short: "Creates or updates the compose.yaml file based on the current Score file.",
		Long: `Creates or updates the compose.yaml file based on the current Score file.

The compose.yaml file will be located by default in the same directory as the
.score-compose directory. If working with multiple score files for different
workloads, as long as they are in subdirectories, score-compose generate can
be used to add them to the current compose.yaml file.
`,
		Example: `# Add the current score.yaml to the current compose.yaml
` + CLIName + ` generate

# Add the current score.yaml to the current compose.yaml overriding an environment variable
` + CLIName + ` generate -p containers.my-container.variables.MESSAGE='"Hello World!"'
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Generate(scoreFilePath, overridesFile, propertyOverrides, buildStr)
		},
		Args: cobra.ExactArgs(0),
	}

	generateCmd.Flags().StringVarP(&scoreFilePath, "file", "f", ScoreFilePathDefault, "The path to the Score file.")
	generateCmd.Flags().StringVar(&overridesFile, "overrides", "", "Overrides properties in the score file using a JSON Merge methodology.")
	generateCmd.Flags().StringArrayVarP(&propertyOverrides, "property", "p", nil, "Overrides a property selected property value. Properties are defined JSONPaths and the values are parsed as YAML. Omitting the \"=\" and value causes the property to be removed. Otherwise the property is added or updated.")
	generateCmd.Flags().StringVar(&buildStr, "build", ".", "The value for the build field in the service when image is set to \".\". The value will be parsed as YAML.")
	return generateCmd
}

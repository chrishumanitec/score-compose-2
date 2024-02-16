package config

import (
	"os"
	"path"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {

	// The directory that contains the context files for score-compose
	ContextDir string `json:"contextDir"`

	// The directory that the compose.yaml file is in i.e. the current directory that everything should be relative to
	CurrentDir string `json:"currentDir"`
}

func loadContext(config *Config) {
	// if ContextDir is empty, start with the current directory and traverse up
	// the filesystem tree until we find a directory called .score-compose
}

func findContextDir(currentDir string) string {
	if currentDir == "/" || currentDir == "." {
		return ""
	}
	if info, err := os.Stat(path.Join(currentDir, ".score-compose")); err != nil {
		if os.IsNotExist(err) {
			return findContextDir(path.Dir(currentDir))
		}
		return ""
	} else {
		if info.IsDir() {
			return path.Join(currentDir, ".score-compose")
		}
	}
	return ""
}

// Loads configuration values.
func Load(config *Config, flagSet *pflag.FlagSet) error {

	if currentDir, err := os.Getwd(); err == nil {

		contextDir := findContextDir(currentDir)
		if contextDir == "" {
			contextDir = path.Join(currentDir, ".score-compose")
		}
		viper.SetDefault("contextDir", contextDir)
	}

	viper.SetEnvPrefix("SCORE_COMPOSE")
	viper.BindEnv("contextDir", "SCORE_COMPOSE_CONTEXT_DIR") // We need to manually set the prefix here.
	viper.BindPFlag("contextDir", flagSet.Lookup("context-dir"))

	if err := viper.Unmarshal(&config); err != nil {
		return err
	}
	loadContext(config)
	return nil
}

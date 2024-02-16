package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"

	"github.com/score-spec/score-compose/internal/config"
	"github.com/score-spec/score-compose/internal/resources"
	"github.com/score-spec/score-compose/internal/score"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

var projectNameRegexp = regexp.MustCompile("^[a-z0-9][a-z0-9_-]*$")

func WriteAsYAMLFile(obj any, path string, fs afero.Fs) error {
	file, err := fs.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := yaml.NewEncoder(file).Encode(obj); err != nil {
		return err
	}
	return nil
}

type Commands interface {
	Init(workloadName string) error
	Generate(scoreFilePath, overrideFile string, propertyOverrides []string, build string) error
}

type commands struct {
	config *config.Config
	appFs  afero.Fs
}

func New(conf *config.Config, appFs afero.Fs) Commands {
	return &commands{
		config: conf,
		appFs:  appFs,
	}
}

func (c *commands) Init(composeProject string) error {
	if !projectNameRegexp.MatchString(composeProject) {
		return fmt.Errorf("not a valid compose project id: %s", composeProject)
	}
	if err := c.appFs.MkdirAll(c.config.ContextDir, 0o777); err != nil {
		return fmt.Errorf("unable to create context directory \"%s\": %w", c.config.ContextDir, err)
	}
	scoreContext, err := score.LoadContext(c.config.ContextDir, c.appFs)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("unable to read from context directory \"%s\": %w", c.config.ContextDir, err)
		}
	} else {
		return fmt.Errorf("there is already a score compose project at this location: %s", c.config.ContextDir)
	}

	scoreContext.ComposeProjectName = composeProject
	if err = scoreContext.WriteOut(); err != nil {
		return fmt.Errorf("unable to write file to context directory \"%s\": %w", c.config.ContextDir, err)
	}
	return nil
}

func (c *commands) Generate(scoreFilePath, overrideFilePath string, propertyOverrides []string, build string) error {
	if overrideFilePath != "" {
		overrideFile, err := c.appFs.Open(overrideFilePath)
		if err != nil {
			return fmt.Errorf("cannot open override file \"%s\": %w", overrideFilePath, err)
		}
		defer overrideFile.Close()
		return errors.New("override files not supported")
	}
	scoreFile, err := c.appFs.Open(scoreFilePath)
	if err != nil {
		return fmt.Errorf("unable to open score file \"%s\": %w", scoreFilePath, err)
	}
	defer scoreFile.Close()
	spec, err := score.LoadSpec(scoreFile, propertyOverrides)
	if err != nil {
		return err
	}

	scoreContext, err := score.LoadContext(c.config.ContextDir, c.appFs)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("no context directory found. Have you run score-compose init?")
		}
		return err
	}

	if err := scoreContext.Update(spec); err != nil {
		return err
	}

	fullFilesDir := path.Join(c.config.ContextDir, "files")
	relFilesDir := path.Join(path.Base(c.config.ContextDir), "files")

	fullVolumesDir := path.Join(c.config.ContextDir, "volumes")
	relVolumesDir := path.Join(path.Base(c.config.ContextDir), "volumes")

	// Provision resources
	provisionerYAML, err := resources.ProvisionerYAML.Open("provisioners.yaml")
	if err != nil {
		return err
	}
	provisioner, err := resources.LoadProvisioners(provisionerYAML, map[string]string{
		"files":   relFilesDir,
		"volumes": relVolumesDir,
	})
	if err != nil {
		return err
	}

	if err := scoreContext.ProvisionResources(provisioner); err != nil {
		return fmt.Errorf("provisioning resources: %w", err)
	}

	if err := scoreContext.ProvisionWorkloads(); err != nil {
		return fmt.Errorf("provisioning workloads: %w", err)
	}

	composeProject, files, volumeDirs, err := scoreContext.GenerateComposeProject()
	if err != nil {
		return fmt.Errorf("generating compose file: %w", err)
	}

	c.appFs.MkdirAll(fullFilesDir, 0o777)

	for fileName, fileContent := range files {
		filePath := path.Join(fullFilesDir, fileName)
		c.appFs.MkdirAll(path.Dir(filePath), 0o777)
		file, err := c.appFs.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0o664)
		if err != nil {
			return fmt.Errorf("unable to write file %s: %w", filePath, err)
		}
		_, err = file.WriteString(fileContent)
		file.Close()
		if err != nil {
			return fmt.Errorf("unable to write file %s: %w", filePath, err)
		}
	}

	c.appFs.MkdirAll(fullVolumesDir, 0o777)
	for volumeDir := range volumeDirs {
		volumeDirPath := path.Join(fullVolumesDir, volumeDir)
		c.appFs.MkdirAll(path.Dir(volumeDirPath), 0o777)
	}

	rawComposeFilePath := path.Join(c.config.CurrentDir, "compose.yaml")
	if err := WriteAsYAMLFile(composeProject, rawComposeFilePath, c.appFs); err != nil {
		return fmt.Errorf("unable to write compose file %s: %w", rawComposeFilePath, err)
	}

	return scoreContext.WriteOut()
}

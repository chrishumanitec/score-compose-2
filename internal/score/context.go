package score

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/score-spec/score-compose/internal/resources"
	"github.com/score-spec/score-compose/internal/version"
	score "github.com/score-spec/score-go/types"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

func LoadContext(contextDir string, appFs afero.Fs) (*Context, error) {
	context := NewContext(contextDir, appFs)

	contextFile, err := appFs.Open(path.Join(contextDir, "context.yaml"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return context, fs.ErrNotExist
		}
		return nil, err
	}
	defer contextFile.Close()

	if err := yaml.NewDecoder(contextFile).Decode(context); err != nil {
		return nil, err
	}
	return context, nil
}

func (c *Context) WriteOut() error {
	if err := c.appFs.MkdirAll(c.contextDir, 0777); err != nil {
		return err
	}

	contextFile, err := c.appFs.OpenFile(path.Join(c.contextDir, "context.yaml"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer contextFile.Close()
	if err := yaml.NewEncoder(contextFile).Encode(c); err != nil {
		return err
	}
	return nil
}

func NewContext(contextDir string, appFs afero.Fs) *Context {
	return &Context{
		ResourceState: ResourceState{
			Workloads:   map[string]map[string]string{},
			Resources:   map[string]*Resource{},
			Provisioned: map[string]*resources.Provisioned{},
		},
		WorkloadState: WorkloadState{
			Specs:       map[string]*score.Workload{},
			Provisioned: map[string]*resources.Provisioned{},
			Build:       map[string]any{},
		},
		Version:    version.Version,
		appFs:      appFs,
		contextDir: contextDir,
	}
}

func defaultIfNil(str *string) string {
	if str == nil {
		return "default"
	}
	return *str
}

// getGlobalResourceName returns the resource name that is global to the
// context. There could be different instances of resources with the same
// global name - i.e. if they have different types or classes.
func getGlobalResourceName(workloadName, resName string, res score.Resource) string {
	return fmt.Sprintf("%s.%s", workloadName, resName)
}

// getGlobalResourceName returns a unique ID for the resource - representing a
// single instance of the resource.
func getResourceID(resID string, res score.Resource) string {
	return fmt.Sprintf("%s::%s::%s", res.Type, defaultIfNil(res.Class), resID)
}

func (c *Context) Update(workloadSpec *score.Workload) error {
	workloadName, ok := workloadSpec.Metadata["name"].(string)
	if !ok {
		return fmt.Errorf("workload metadata does not have a name property")
	}

	// TODO:
	// - Check if the score file is being updated and the path to the score
	//   file has changed. This should be a warning that there might be 2
	//   score files with the same name being added to the context.
	c.WorkloadState.Specs[workloadName] = workloadSpec

	if _, exists := c.ResourceState.Workloads[workloadName]; !exists {
		c.ResourceState.Workloads[workloadName] = map[string]string{}
	}

	for localResName, res := range workloadSpec.Resources {
		resName := getGlobalResourceName(workloadName, localResName, res)
		resID := getResourceID(resName, res)

		c.ResourceState.Workloads[workloadName][localResName] = resID

		c.ResourceState.Resources[resID] = &Resource{
			ID:       resID,
			Current:  true,
			Name:     resName,
			Resource: res,
		}
	}

	// Update "Current" flag on resources
	currentResourceIDs := map[string]bool{}
	for _, resources := range c.ResourceState.Workloads {
		for _, id := range resources {
			currentResourceIDs[id] = true
		}
	}
	for resID := range c.ResourceState.Resources {
		c.ResourceState.Resources[resID].Current = currentResourceIDs[resID]
	}

	return nil
}

package score

import (
	"github.com/score-spec/score-compose/internal/resources"
	"github.com/spf13/afero"

	score "github.com/score-spec/score-go/types"
)

type Context struct {
	ResourceState ResourceState `json:"resources,omitempty"`
	WorkloadState WorkloadState `json:"workloads,omitempty"`

	Version            string `json:"version"`
	ComposeProjectName string `json:"composeProjectName"`

	appFs      afero.Fs `json:"-"`
	contextDir string   `json:"-"`
}

type ResourceState struct {
	// For each workload, a map of workload names to the ID
	Workloads map[string]map[string]string `json:"workloads"`

	// A map of IDs -> resource objects
	Resources map[string]*Resource `json:"resources"`

	Provisioned map[string]*resources.Provisioned `json:"provisioned"`
}

type Resource struct {
	// This is a unique ID of the resource made up of type, class and name
	ID string `json:"id"`
	// Current is true if it is in Context.ResourceState.Workloads
	Current bool `json:"current"`
	// This is either <workload-id>.<name> or external.<external-id>
	Name string `json:"name"`
	score.Resource
}

type WorkloadState struct {
	// The score specs that have been added to the context
	Specs map[string]*score.Workload `json:"specs"`

	Provisioned map[string]*resources.Provisioned `json:"provisioned"`

	Build map[string]any `json:"build"`
}

package resources

type Paths struct {
	Files   string `json:"files"`
	Volumes string `json:"volumes"`
}

type ProvisionerDefinitions struct {
	Types map[string]*ProvisionerTemplates `json:"types"`
	Paths map[string]string                `json:"paths"`
}

type ProvisionerTemplates struct {
	Init       string `json:"init"`
	Outputs    string `json:"outputs"`
	Services   string `json:"services"`
	Networks   string `json:"networks"`
	Files      string `json:"files"`
	State      string `json:"state"`
	Shared     string `json:"shared"`
	VolumeDirs string `json:"volumeDirs"`
}

type Provisioned struct {
	// Services section in the Compose File
	Services map[string]any `json:"services"`
	// Network section in the Compose File
	Networks map[string]any `json:"networks"`

	// State that will be passed back into the provisioning on the next step
	State map[string]any `json:"state"`

	// Files that can be used to gather things together
	Files map[string]any `json:"files"`

	// VolumeDirs directories that will back volumes
	VolumeDirs map[string]any `json:"volumeDirs"`

	// Outputs that will be used in placeholders
	Outputs map[string]any `json:"outputs"`
}

type Provisioner interface {
	Provision(id, resType, resClass string, params, state, shared map[string]any) (*Provisioned, error)
}

type Provisioners interface {
}

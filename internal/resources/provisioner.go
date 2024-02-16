package resources

import (
	"fmt"
	"io"

	"embed"

	"gopkg.in/yaml.v3"
)

//go:embed provisioners.yaml
var ProvisionerYAML embed.FS

func (p *ProvisionerDefinitions) Provision(id, resType, resClass string, params, state, shared map[string]any) (*Provisioned, error) {
	provisioner, exists := p.Types[resType]
	if !exists {
		return nil, fmt.Errorf("no implementation for type %s", resType)
	}
	return provisioner.Provision(id, resType, resClass, params, state, shared, p.Paths)
}

func LoadProvisioners(provisionerYAML io.Reader, paths map[string]string) (*ProvisionerDefinitions, error) {
	provisioners := ProvisionerDefinitions{}
	err := yaml.NewDecoder(provisionerYAML).Decode(&provisioners.Types)
	if err != nil {
		return nil, err
	}
	provisioners.Paths = paths
	return &provisioners, nil
}

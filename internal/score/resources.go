package score

import (
	"github.com/score-spec/score-compose/internal/resources"
)

func (c *Context) ProvisionResources(provisioner resources.Provisioner) error {
	sharedState := map[string]any{}
	for id, res := range c.ResourceState.Resources {
		existingState := map[string]any{}
		if p, exists := c.ResourceState.Provisioned[id]; exists {
			existingState = p.State
		}
		class := defaultIfNil(res.Class)
		params := map[string]any{}
		if res.Params != nil {
			params = map[string]any(res.Params)
		}
		var err error
		c.ResourceState.Provisioned[id], err = provisioner.Provision(id, res.Type, class, params, existingState, sharedState)
		if err != nil {
			return err
		}
	}
	return nil
}

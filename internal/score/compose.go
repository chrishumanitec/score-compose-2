package score

import (
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/compose-spec/compose-go/v2/loader"
	compose "github.com/compose-spec/compose-go/v2/types"
	"github.com/score-spec/score-compose/internal/resources"
	"gopkg.in/yaml.v3"
)

func dumpYAML(o any) string {
	yamlBytes, err := yaml.Marshal(o)
	if err != nil {
		return fmt.Sprintf("dumping-yaml: %v", err)
	}
	return string(yamlBytes)
}

func isMap(a any) bool {
	_, isMap := a.(map[string]any)
	return isMap
}

// mergeMap performs an additive merge. An error is thrown if any non-map values do not match.
// nil values are ignored.
func mergeMap(a, b map[string]any, path string) (map[string]any, error) {
	out := map[string]any{}
	for k := range a {
		if _, exists := b[k]; exists {
			if isMap(a[k]) && isMap(b[k]) {
				var err error
				out[k], err = mergeMap(a[k].(map[string]any), b[k].(map[string]any), path+asJsonPathSegment(k))
				if err != nil {
					return nil, err
				}
			} else if a[k] == nil {
				out[k] = b[k]
			} else if b[k] == nil {
				out[k] = a[k]
			} else if reflect.DeepEqual(a[k], b[k]) {
				out[k] = a[k]
			} else {
				return nil, fmt.Errorf("conflict when merging maps for key %s", path+asJsonPathSegment(k))
			}
		} else {
			out[k] = a[k]
		}
	}
	for k := range b {
		if _, exists := a[k]; !exists {
			out[k] = b[k]
		}
	}
	return out, nil
}

func escapeComposeInterpolation(o any) any {
	switch a := o.(type) {
	case map[string]any:
		for k, v := range a {
			a[k] = escapeComposeInterpolation(v)
		}
	case []any:
		for i, v := range a {
			a[i] = escapeComposeInterpolation(v)
		}
	case string:
		o = strings.ReplaceAll(a, "$", "$$")
	}
	return o
}

func Collate(provisionedMap map[string]*resources.Provisioned, selector func(*resources.Provisioned) map[string]any, name string) (map[string]any, error) {
	collated := map[string]any{}
	var err error
	for id, provisioned := range provisionedMap {
		collated, err = mergeMap(selector(provisioned), collated, name+asJsonPathSegment(id))
		if err != nil {
			return nil, err
		}
	}
	return collated, nil
}

func selectServices(p *resources.Provisioned) map[string]any {
	return p.Services
}

func selectFiles(p *resources.Provisioned) map[string]any {
	return p.Files
}

func selectNetworks(p *resources.Provisioned) map[string]any {
	return p.Networks
}

func selectVolumeDirs(p *resources.Provisioned) map[string]any {
	return p.VolumeDirs
}

func withSkipInterpolation(o *loader.Options) {
	o.SkipInterpolation = true
}

func (c *Context) GenerateComposeProject() (map[string]any, map[string]string, map[string]any, error) {

	// Collate resources
	resourceServices, err := Collate(c.ResourceState.Provisioned, selectServices, "services")
	if err != nil {
		return nil, nil, nil, err
	}
	resourceFiles, err := Collate(c.ResourceState.Provisioned, selectFiles, "files")
	if err != nil {
		return nil, nil, nil, err
	}
	resourceNetworks, err := Collate(c.ResourceState.Provisioned, selectNetworks, "networks")
	if err != nil {
		return nil, nil, nil, err
	}
	resourceVolumeDirs, err := Collate(c.WorkloadState.Provisioned, selectVolumeDirs, "volumeDirs")
	if err != nil {
		return nil, nil, nil, err
	}

	// Collate workloads
	workloadServices, err := Collate(c.WorkloadState.Provisioned, selectServices, "services")
	if err != nil {
		return nil, nil, nil, err
	}
	workloadFiles, err := Collate(c.WorkloadState.Provisioned, selectFiles, "files")
	if err != nil {
		return nil, nil, nil, err
	}
	workloadNetworks, err := Collate(c.WorkloadState.Provisioned, selectNetworks, "networks")
	if err != nil {
		return nil, nil, nil, err
	}
	workloadVolumeDirs, err := Collate(c.WorkloadState.Provisioned, selectVolumeDirs, "volumeDirs")
	if err != nil {
		return nil, nil, nil, err
	}

	mergedServices, err := mergeMap(workloadServices, resourceServices, "services")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generating compose services: merging services: %w", err)
	}

	mergedNetworks, err := mergeMap(workloadNetworks, resourceNetworks, "networks")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generating compose services: merging services: %w", err)
	}

	mergedFiles, err := mergeMap(workloadFiles, resourceFiles, "files")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generating compose volumes: merging files: %w", err)
	}
	outFiles := map[string]string{}
	for fn, content := range mergedFiles {
		strContent, ok := content.(string)
		if !ok {
			return nil, nil, nil, fmt.Errorf("content of file %s is not a string: got %T", fn, content)
		}
		outFiles[fn] = strContent
	}

	mergedVolumeDirs, err := mergeMap(workloadVolumeDirs, resourceVolumeDirs, "volumeDirs")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generating compose volumes: merging volumes: %w", err)
	}

	composeProjectMap := map[string]any{
		"name":     "score-compose",
		"services": mergedServices,
		"networks": mergedNetworks,
	}

	composeProjectMap, err = FullyMappify(composeProjectMap)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generating compose file: %w", err)
	}
	composeProjectMap = escapeComposeInterpolation(composeProjectMap).(map[string]any)

	composeYAML, err := ObjToYAMLBytes(composeProjectMap)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generating compose project: converting to YAML: %w", err)
	}

	_, err = loader.Load(compose.ConfigDetails{
		Version:    "3.8",
		WorkingDir: path.Dir(c.contextDir),
		ConfigFiles: []compose.ConfigFile{
			{
				Filename: "compose.yaml",
				Content:  composeYAML,
				Config:   nil,
			},
		},
		Environment: nil,
	}, withSkipInterpolation)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generating compose project: validating compose file: %w", err)
	}

	/*
		project := compose.Project{
			Name:              "",
			WorkingDir:        "",
			Services:          services,
			Networks:          networks,
			Volumes:           map[string]compose.VolumeConfig{},
			Secrets:           map[string]compose.SecretConfig{},
			Configs:           map[string]compose.ConfigObjConfig{},
			Extensions:        map[string]interface{}{},
			IncludeReferences: map[string][]compose.IncludeConfig{},
			ComposeFiles:      []string{},
			Environment:       map[string]string{},
			DisabledServices:  map[string]compose.ServiceConfig{},
			Profiles:          []string{},
		}*/

	return composeProjectMap, outFiles, mergedVolumeDirs, nil
}

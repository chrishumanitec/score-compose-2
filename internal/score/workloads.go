package score

import (
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"sort"
	"strings"

	compose "github.com/compose-spec/compose-go/v2/types"
	"github.com/score-spec/score-compose/internal/resources"
	score "github.com/score-spec/score-go/types"
)

var (
	placeholderRegexp = regexp.MustCompile(`\${[a-zA-Z0-9.\-_]+}`)
)

func selectByPath(path []string, val any) (string, error) {
	if len(path) > 1 {
		if valMap, ok := val.(map[string]any); ok {
			if nextVal, exists := valMap[path[0]]; exists {
				return selectByPath(path[1:], nextVal)
			}
		}
		return "", fmt.Errorf("unable to resolve path .%s in type %T", strings.Join(path, "."), val)
	}
	return fmt.Sprintf("%v", val), nil
}

func (c *Context) placeholderValue(workloadName, placeholder string) (string, error) {
	parts := strings.Split(placeholder, ".")
	if len(parts) > 2 {
		switch parts[0] {
		case "resources":
			if w, exists := c.ResourceState.Workloads[workloadName]; exists {
				if id, exists := w[parts[1]]; exists {
					if provisioned, exists := c.ResourceState.Provisioned[id]; exists {
						if output, exists := provisioned.Outputs[parts[2]]; exists {
							return selectByPath(parts[3:], output)
						}
					}
				}
			}
		}
	}
	return "", fmt.Errorf("unable to resolve placeholder \"%s\"", placeholder)
}

func (c *Context) replacePlaceholderInString(workloadName, str string) (string, error) {
	for _, match := range placeholderRegexp.FindAllString(str, -1) {
		placeholder := strings.TrimSuffix(match[2:], "}")
		replacementValue, err := c.placeholderValue(workloadName, placeholder)
		if err != nil {
			return "", err
		}
		str = strings.ReplaceAll(str, match, replacementValue)
	}
	return str, nil
}

func (c *Context) ProvisionWorkloads() error {
	for id, spec := range c.WorkloadState.Specs {
		var err error
		c.WorkloadState.Provisioned[id], err = c.ProvisionWorkload(spec)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
	func (c *Context) ProvisionWorkload(spec *score.Workload) (*resources.Provisioned, error) {
		p := resources.Provisioned{
			Services: map[string]any{},
			Networks: map[string]any{},
			State:    map[string]any{},
			Files:    map[string]any{},
			Outputs:  map[string]any{},
		}

		for containerID, container := range spec.Containers {
			service := map[string]any{}
			p.Services[workloadName+"-"+containerID] = service

			if len(container.Variables) > 0 {
				environment := map[string]any{}
				service["environment"] = environment
				for k, v := range container.Variables {
					var err error
					environment[k], err = c.replacePlaceholderInString(workloadName, v)
					if err != nil {
						return nil, fmt.Errorf("in workload %s, containers.%s.variables.%s: %w", workloadName, containerID, k, err)
					}
				}

			}

			if container.Image == "." {
				service["build"] = c.WorkloadState.Build[workloadName]
			}
			/*
				if len(container.Files) > 0 {
					volumes := map[string]any{}
					service["volumes"] = volumes
					for k, file := range container.Files {
						if file.NoExpand != nil && !*file.NoExpand {
							var err error
							if content, ok := file.Content.(string); ok {
								file.Content, err = c.replacePlaceholderInString(workloadName, content)
								if err != nil {
									return fmt.Errorf("in workload %s, containers.%s.files.%s: %w", workloadName, containerID, k)
								}
							} else {
								fmt.Errorf("in workload %s, containers.%s.files.%s: do not support legacy array format", workloadName, containerID, k)
							}
						}
					}
				}*
		}
		return &p, nil
	}
*/

// ProvisionWorkload generates a parts of a compose file that can be collated together later.
// This is lifted from https://github.com/score-spec/score-compose/blob/c507418d423811ce91685e4013784983b5b52847/internal/compose/convert.go#L20 with a few changes
func (c *Context) ProvisionWorkload(spec *score.Workload) (*resources.Provisioned, error) {
	var workloadName string
	var ok bool
	if workloadName, ok = spec.Metadata["name"].(string); !ok {
		return nil, errors.New("workload metadata is missing a name")
	}

	if len(spec.Containers) == 0 {
		return nil, errors.New("workload does not have any containers to convert into a compose service")
	}

	p := &resources.Provisioned{
		Services: map[string]any{},
		Networks: map[string]any{},
		State:    map[string]any{},
		Files:    map[string]any{},
		Outputs:  map[string]any{},
	}

	var ports []compose.ServicePortConfig
	if spec.Service != nil && len(spec.Service.Ports) > 0 {
		ports = []compose.ServicePortConfig{}
		for _, pSpec := range spec.Service.Ports {
			var pubPort = fmt.Sprintf("%v", pSpec.Port)
			ports = append(ports, compose.ServicePortConfig{
				Published: pubPort,
				Target:    uint32(DerefOr(pSpec.TargetPort, pSpec.Port)),
				Protocol:  DerefOr(pSpec.Protocol, ""),
			})
		}
	}

	// When multiple containers are specified we need to identify one container as the "main" container which will own
	// the network and use the native workload name. All other containers in this workload will have the container
	// name appended as a suffix. We use the natural sort order of the container names and pick the first one
	containerNames := make([]string, 0, len(spec.Containers))
	for name := range spec.Containers {
		containerNames = append(containerNames, name)
	}
	sort.Strings(containerNames)

	firstContainerInService := workloadName + "-" + containerNames[0]

	for _, containerName := range containerNames {
		cSpec := spec.Containers[containerName]

		var env = make(compose.MappingWithEquals, len(cSpec.Variables))
		for key, val := range cSpec.Variables {
			resolvedVar, err := c.replacePlaceholderInString(workloadName, val)
			if err != nil {
				return nil, fmt.Errorf("resolving placeholders .%s.containers.%s.variables.%s: %w", workloadName, containerName, key, err)
			}
			env[key] = &resolvedVar
		}

		// NOTE: Sorting is necessary for DeepEqual call within our Unit Tests to work reliably
		sort.Slice(ports, func(i, j int) bool {
			return ports[i].Published < ports[j].Published
		})
		// END (NOTE)

		volumes := make([]compose.ServiceVolumeConfig, 0, len(cSpec.Volumes)+len(cSpec.Files))
		for idx, vol := range cSpec.Volumes {
			if vol.Path != nil && *vol.Path != "" {
				return nil, fmt.Errorf("can't mount named volume with sub path '%s': %w", *vol.Path, errors.New("not supported"))
			}
			resolvedSource, err := c.replacePlaceholderInString(workloadName, vol.Source)
			if err != nil {
				return nil, fmt.Errorf("resolving placeholders .%s.containers.%s.volumes.%d: %w", workloadName, containerName, idx, err)
			}
			volumes = append(volumes, compose.ServiceVolumeConfig{
				Type:     "volume",
				Source:   resolvedSource,
				Target:   vol.Target,
				ReadOnly: DerefOr(vol.ReadOnly, false),
			})
		}
		for idx, file := range cSpec.Files {
			var content string
			if file.Source != nil {
				// read the file
				contentFile, err := c.appFs.Open(*file.Source)
				if err != nil {
					return nil, fmt.Errorf("loading source file .%s.containers.%s.files.%d (%s): unable to read file %s: %w", workloadName, containerName, idx, file.Target, *file.Source, err)
				}
				contentBytes, err := io.ReadAll(contentFile)
				if err != nil {
					return nil, fmt.Errorf("loading source file .%s.containers.%s.files.%d (%s): unable to read file %s: %w", workloadName, containerName, idx, file.Target, *file.Source, err)
				}
				content = string(contentBytes)
			} else {
				var ok bool
				content, ok = file.Content.(string)

				if !ok {
					contentArray, ok := file.Content.([]string)
					if !ok {
						return nil, fmt.Errorf("parsing content of .%s.containers.%s.files.%d (%s): expected string, got %T", workloadName, containerName, idx, file.Target, file.Content)
					}
					content = strings.Join(contentArray, "\n")
				}
			}
			if file.NoExpand == nil || !*file.NoExpand {
				var err error
				content, err = c.replacePlaceholderInString(workloadName, content)
				if err != nil {
					return nil, fmt.Errorf("resolving placeholders .%s.containers.%s.files.%d (%s): %w", workloadName, containerName, idx, file.Target, err)
				}
			}

			sourceFileName := HashOfString(content)
			p.Files[sourceFileName] = content
			volumes = append(volumes, compose.ServiceVolumeConfig{
				Type:     "bind",
				Source:   path.Join(c.contextDir, "files", sourceFileName),
				Target:   file.Target,
				ReadOnly: true,
			})
		}
		// NOTE: Sorting is necessary for DeepEqual call within our Unit Tests to work reliably
		sort.Slice(volumes, func(i, j int) bool {
			return volumes[i].Target < volumes[j].Target
		})
		// END (NOTE)

		var svc = compose.ServiceConfig{
			Name:        workloadName + "-" + containerName,
			Image:       cSpec.Image,
			Entrypoint:  cSpec.Command,
			Command:     cSpec.Args,
			Environment: env,
			Ports:       ports,
			Volumes:     volumes,
		}

		// if we are not the "first" service, then inherit the network from the first service
		if len(p.Services) > 0 {
			svc.Ports = nil
			svc.NetworkMode = "service:" + firstContainerInService

		}
		p.Services[svc.Name] = svc
	}
	return p, nil
}

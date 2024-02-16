package resources

// Templates are go templates and must evaluate to YAML objects.
//
// The go-templates have broadly the same functions available as Helm.
//
// Templates are evaluated in the row order of the following table. Each
// template has the following set of inputs:
//
//  .id     - (string) a unique ID for the resource that is global to the
//                     context. It is guaranteed to be a RFC 1123 Label
//                     Name; https://tools.ietf.org/html/rfc1123
//  .class  - (string) the class of the resource
//  .type   - (string) the type of the resource
//  .paths  - (map)    a map of paths useful for working with volumes.
//                     files -   the directory into which the outputted files
//                               are written.
//                     volumes - the directory where volumes directories are
//                               created.
//  .params - (map)    any resource input parameters from the score file
//  .state  - (map)    the last state that was stored for this resource
//  .global - (map)    the current global state
//
// Additional inputs depend on the template as enumerated in this table.
//
// | tpl \ inputs | .init | .outputs | .services | .files | .volumeDirs |
// | ------------ | ----- | -------- | --------- | ------ | ----------- |
// | init         |       |          |           |        |             |
// | outputs      |   x   |          |           |        |             |
// | files        |   x   |     x    |           |        |             |
// | networks     |   x   |     x    |           |        |             |
// | service      |   x   |     x    |           |        |             |
// | volumes      |   x   |     x    |           |        |             |
// | shared       |   x   |     x    |     x     |   x    |      x      |
// | state        |   x   |     x    |     x     |   x    |      x      |

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/score-spec/score-compose/internal/util"
	"gopkg.in/yaml.v3"
)

// HashOfString returns the hash of a string encoded as a hex string
// It is intended to be safe for use as a filename on all platforms
func HashOfString(str string) string {
	hash := sha1.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

// processTemplate applies a go template to the provided template with data.
//
// path is used to generate useful error messages.
func processTemplate(tmpl string, data map[string]any, path string) (map[string]any, error) {

	f := sprig.GenericFuncMap()

	goTmpl, err := template.New(path).Funcs(f).Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("error parsing template %s: %w", path, err)
	}

	var yamlBuffer bytes.Buffer
	err = goTmpl.Execute(&yamlBuffer, data)
	if err != nil {
		return nil, fmt.Errorf("error executing template %s: %w", path, err)
	}

	var out map[string]any
	err = yaml.Unmarshal(yamlBuffer.Bytes(), &out)
	if err != nil {
		fmt.Println(yamlBuffer.String())
		return nil, fmt.Errorf("output of template %s was not valid YAML: %w", path, err)
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

func (t *ProvisionerTemplates) Provision(id, resType, resClass string, params, state, shared map[string]any, paths map[string]string) (*Provisioned, error) {
	if params == nil {
		params = map[string]any{}
	}
	if state == nil {
		state = map[string]any{}
	}
	if shared == nil {
		shared = map[string]any{}
	}

	id = HashOfString(id)

	inputs := map[string]any{
		"id":     id,
		"type":   resType,
		"class":  resClass,
		"params": params,
		"paths":  paths,
		"state":  state,
		"shared": shared,
	}
	var err error

	init, err := processTemplate(t.Init, inputs, "init")
	if err != nil {
		return nil, err
	}
	inputs["init"] = init

	outputs, err := processTemplate(t.Outputs, inputs, "outputs")
	if err != nil {
		return nil, err
	}
	inputs["outputs"] = outputs

	services, err := processTemplate(t.Services, inputs, "services")
	if err != nil {
		return nil, err
	}

	networks, err := processTemplate(t.Networks, inputs, "networks")
	if err != nil {
		return nil, err
	}

	files, err := processTemplate(t.Files, inputs, "files")
	if err != nil {
		return nil, err
	}
	volumeDirs, err := processTemplate(t.VolumeDirs, inputs, "volumeDirs")
	if err != nil {
		return nil, err
	}

	inputs["services"] = services
	inputs["networks"] = networks
	inputs["files"] = files
	inputs["volumeDirs"] = volumeDirs
	state, err = processTemplate(t.State, inputs, "state")
	if err != nil {
		return nil, err
	}

	if sharedUpdate, err := processTemplate(t.Shared, inputs, "shared"); err != nil {
		return nil, err
	} else {
		util.JsonMerge(shared, sharedUpdate)
	}

	return &Provisioned{
		Services:   services,
		Networks:   networks,
		State:      state,
		Files:      files,
		VolumeDirs: volumeDirs,
		Outputs:    outputs,
	}, nil
}

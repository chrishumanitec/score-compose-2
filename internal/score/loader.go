package score

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	loader "github.com/score-spec/score-go/loader"
	schema "github.com/score-spec/score-go/schema"
	score "github.com/score-spec/score-go/types"
	"github.com/tidwall/sjson"
	"gopkg.in/yaml.v3"
)

/*
// overrideValueInMap performs in-place overrides for the value at the nested index

	func overrideValueInMap(in map[string]any, path []string, pathIndex int, value any) (any, error) {
		if len(path) == 0 {
			return in, fmt.Errorf("path is empty")
		}
		var err error
		if len(path)-pathIndex > 1 {
			switch v := in[path[pathIndex]].(type) {
			case map[string]any:
				in[path[pathIndex]], err = overrideValueInMap(v, path, pathIndex+1, value)
			case []any:
				in[path[pathIndex]], err = overrideValueInSlice(v, path, pathIndex+1, value)
			default:
				err = fmt.Errorf("path does not resolve: %s", strings.Join(path[:pathIndex], "."))
			}
			return in, err
		}
		if value == nil {
			delete(in, path[pathIndex])
		} else {
			in[path[pathIndex]] = value
		}
		return in, nil

}

// overrideValueInMap performs in-place overrides for the value at the nested index

	func overrideValueInSlice(in []any, path []string, pathIndex int, value any) (any, error) {
		if len(path) == 0 {
			return in, fmt.Errorf("path is empty")
		}
		if index, err := strconv.Atoi(path[0]); err != nil {
			if len(in) <= index || index < 0 {
				return in, fmt.Errorf("path involves out of bounds index %s", strings.Join(path[:pathIndex], "."))
			}

			if len(path)-pathIndex > 1 {
				switch v := in[index].(type) {
				case map[string]any:
					in[index], err = overrideValueInMap(v, path, pathIndex+1, value)
				case []any:
					in[index], err = overrideValueInSlice(v, path, pathIndex+1, value)
				default:
					err = fmt.Errorf("path does not resolve: %s", strings.Join(path[:pathIndex], "."))
				}
				return in, err
			}
			if value == nil {
				in = append(in[:index], path[index+1:])
			} else {
				in[index] = value
			}
			return in, nil
		} else {
			return in, fmt.Errorf("cannot index an array with a non-integer \"%s\" for path: %s", path[pathIndex], strings.Join(path[:pathIndex], "."))
		}
	}
*/

// Mainly lifted from:
func LoadSpec(scoreSpecYAML io.Reader, overrides []string) (*score.Workload, error) {
	var err error
	var srcMap map[string]interface{}
	if err = loader.ParseYAML(&srcMap, scoreSpecYAML); err != nil {
		return nil, err
	}

	for _, pstr := range overrides {

		jsonBytes, err := json.Marshal(srcMap)
		if err != nil {
			return nil, fmt.Errorf("marshalling score spec: %w", err)
		}

		pmap := strings.SplitN(pstr, "=", 2)
		if len(pmap) <= 1 {
			var path = pmap[0]
			if jsonBytes, err = sjson.DeleteBytes(jsonBytes, path); err != nil {
				return nil, fmt.Errorf("removing '%s': %w", path, err)
			}
		} else {
			var path = pmap[0]
			var val interface{}
			if err := yaml.Unmarshal([]byte(pmap[1]), &val); err != nil {
				val = pmap[1]
			}

			if jsonBytes, err = sjson.SetBytes(jsonBytes, path, val); err != nil {
				return nil, fmt.Errorf("overriding '%s': %w", path, err)
			}
		}

		if err = json.Unmarshal(jsonBytes, &srcMap); err != nil {
			return nil, fmt.Errorf("unmarshalling score spec: %w", err)
		}
	}

	if err := schema.Validate(srcMap); err != nil {
		return nil, fmt.Errorf("validating workload spec: %w", err)
	}

	var spec score.Workload
	if err = loader.MapSpec(&spec, srcMap); err != nil {
		return nil, fmt.Errorf("validating workload spec: %w", err)
	}
	return &spec, nil
}

package score

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

const (
	PatternID = `^[a-z0-9]+(?:-[a-z0-9]+)*$`
)

var (
	ValidID = regexp.MustCompile(PatternID)
)

// IsValidID returns true if the string is a valid ID.
//
// A valid ID is made up of lowercase alphanumeric characters and a dash "-".
// The ID cannot start
func IsValidID(str string) bool {
	return ValidID.MatchString(str)
}

func Ref[k any](input k) *k {
	return &input
}

func DerefOr[k any](input *k, def k) k {
	if input == nil {
		return def
	}
	return *input
}

func exists(a any) bool {
	_, isMap := a.(map[string]any)
	return isMap
}

func YAMLStringToObj(str string) (any, error) {
	var obj any
	err := yaml.Unmarshal([]byte(str), obj)
	return obj, err
}

func ObjToYAMLBytes(obj any) ([]byte, error) {
	b, err := yaml.Marshal(obj)
	return b, err
}

func SliceYAMLStringsToSliceObjects(str []string) ([]any, error) {
	objects := make([]any, len(str))
	for i, s := range str {
		var err error
		objects[i], err = YAMLStringToObj(s)
		if err != nil {
			return nil, err
		}
	}
	return objects, nil
}

func MapStructureDecoderForJson(o any) *mapstructure.Decoder {
	d, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:               o,
		TagName:              "json",
		IgnoreUntaggedFields: true,
		ZeroFields:           true,
	})
	return d
}

func FullyMappify(m map[string]any) (map[string]any, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	err = json.Unmarshal(b, &out)
	return out, err
}

var javaScriptIdentifier = regexp.MustCompile("^[a-zA-Z_$][0-9a-zA-Z_$]*$")

// asJsonPathSegment is a helper method to add unambiguous path segments.
func asJsonPathSegment(segment string) string {
	if javaScriptIdentifier.MatchString(segment) {
		return "." + segment
	}
	return "[\"" + strings.ReplaceAll(segment, "\"", "\\\"") + "\"]"
}

func existsInMap[T any](m map[string]T, key string) bool {
	_, exists := m["key"]
	return exists
}

// HashOfString returns the hash of a string encoded as a hex string
// It is intended to be safe for use as a filename on all platforms
func HashOfString(str string) string {
	hash := sha1.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

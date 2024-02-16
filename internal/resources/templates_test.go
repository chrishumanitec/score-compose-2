package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvision(t *testing.T) {
	testCases := []struct {
		name     string
		id       string
		resType  string
		resClass string
		params   map[string]any
		state    map[string]any
		shared   map[string]any
		paths    map[string]string

		templates ProvisionerTemplates

		provisioned Provisioned

		expectedShared map[string]any

		shouldSucceed bool
	}{
		{
			name:     "basic resource",
			id:       "test-id",
			resType:  "test-type",
			resClass: "default",
			params:   map[string]any{},
			state:    map[string]any{},
			shared:   map[string]any{},
			paths:    map[string]string{},

			templates: ProvisionerTemplates{
				Init:     "",
				Outputs:  "test: {{ .id }}",
				Services: "test-service:\n  image: \"test-image:latest\"",
				Networks: "",
				Files:    "",
				State:    "",
			},

			provisioned: Provisioned{
				Services: map[string]any{
					"test-service": map[string]any{
						"image": "test-image:latest",
					},
				},
				Networks: map[string]any{},
				State:    map[string]any{},
				Files:    map[string]any{},
				Outputs: map[string]any{
					"test": HashOfString("test-id"),
				},
				VolumeDirs: map[string]any{},
			},

			expectedShared: map[string]any{},

			shouldSucceed: true,
		},
		{
			name:     "complex templates referencing all dependencies",
			id:       "test-id",
			resType:  "test-type",
			resClass: "default",
			params: map[string]any{
				"first": "value",
			},
			state: map[string]any{
				"storedValue": "xyz",
			},
			shared: map[string]any{
				"keepValue":    "keep this",
				"replaceValue": "should have been replaced",
				"updateValue": map[string]any{
					"keep":       "keep this",
					"replace":    "should have been replaced",
					"removeThis": "should not longer exist",
				},
			},
			paths: map[string]string{},

			templates: ProvisionerTemplates{
				Init: `
value: {{ .state.storedValue }}-xyz`,
				Outputs: `
initValue: {{ .init.value }}
stateValue: {{ .state.storedValue }}`,
				Services: `
testservice:
  image: "test-image:{{ .init.value}}"`,
				Networks: `
testnetwork:
  driver: driver-{{ .init.value }}`,
				Files: "testFile: abc",
				State: `
storedValue: {{ .init.value }}
file: {{ .files.testFile }}
output: {{ .outputs.initValue }}
image: {{ .services.testservice.image }}
network-driver: {{ .networks.testnetwork.driver }}`,
				Shared: `
replaceValue: new value
updateValue:
  replace: new value
  removeThis: null`,
				VolumeDirs: `
example/dir: {}
`,
			},

			provisioned: Provisioned{
				Services: map[string]any{
					"testservice": map[string]any{
						"image": "test-image:xyz-xyz",
					},
				},
				Networks: map[string]any{
					"testnetwork": map[string]any{
						"driver": "driver-xyz-xyz",
					},
				},
				Outputs: map[string]any{
					"initValue":  "xyz-xyz",
					"stateValue": "xyz",
				},
				Files: map[string]any{
					"testFile": "abc",
				},
				State: map[string]any{
					"storedValue":    "xyz-xyz",
					"file":           "abc",
					"output":         "xyz-xyz",
					"image":          "test-image:xyz-xyz",
					"network-driver": "driver-xyz-xyz",
				},
				VolumeDirs: map[string]any{
					"example/dir": map[string]any{},
				},
			},

			expectedShared: map[string]any{
				"keepValue":    "keep this",
				"replaceValue": "new value",
				"updateValue": map[string]any{
					"keep":    "keep this",
					"replace": "new value",
				},
			},

			shouldSucceed: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			provisioned, err := testCase.templates.Provision(testCase.id, testCase.resType, testCase.resClass, testCase.params, testCase.state, testCase.shared, testCase.paths)
			if testCase.shouldSucceed {
				require.NoError(t, err)
				assert.Equal(t, testCase.provisioned, *provisioned)
				assert.Equal(t, testCase.expectedShared, testCase.shared)
			}
		})
	}
}

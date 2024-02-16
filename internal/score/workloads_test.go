package score

import (
	"testing"

	compose "github.com/compose-spec/compose-go/v2/types"
	"github.com/score-spec/score-compose/internal/resources"
	score "github.com/score-spec/score-go/types"
	"github.com/stretchr/testify/assert"
)

func TestProvisionWorkload(t *testing.T) {
	testCases := []struct {
		name                string
		workloadSpec        score.Workload
		expectedProvisioned resources.Provisioned
		context             Context
	}{
		{
			name: "simplest workload spec",
			workloadSpec: score.Workload{
				ApiVersion: "spec.score.dev/v1b1",
				Containers: map[string]score.Container{
					"test-container": {
						Image: ".",
					},
				},
				Metadata: map[string]interface{}{
					"name": "test-workload",
				},
				Resources: map[string]score.Resource{},
				Service:   &score.WorkloadService{},
			},
			expectedProvisioned: resources.Provisioned{
				Services: map[string]any{
					"test-workload-test-container": compose.ServiceConfig{
						Name:        "test-workload-test-container",
						Image:       ".",
						Entrypoint:  nil,
						Environment: compose.MappingWithEquals{},
						Command:     nil,
						Ports:       nil,
						Volumes:     []compose.ServiceVolumeConfig{},
					},
				},
				Networks: map[string]any{},
				State:    map[string]any{},
				Files:    map[string]any{},
				Outputs:  map[string]any{},
			},
			context: *NewContext("test", nil),
		},
		{
			name: "workload all features no resources",
			workloadSpec: score.Workload{
				ApiVersion: "spec.score.dev/v1b1",
				Containers: map[string]score.Container{
					"test-container": {
						Args:    []string{"-c", "echo \"Hello Command\""},
						Command: []string{"/bin/sh"},
						Files: []score.ContainerFilesElem{
							{
								Content: "Hello File",
								Target:  "/test/file.txt",
							},
						},
						Image:          ".",
						LivenessProbe:  &score.ContainerProbe{},
						ReadinessProbe: &score.ContainerProbe{},
						Resources:      &score.ContainerResources{},
						Variables: map[string]string{
							"MESSAGE": "Hello Environment",
						},
						Volumes: []score.ContainerVolumesElem{},
					},
				},
				Metadata: map[string]interface{}{
					"name": "test-workload",
				},
				Resources: map[string]score.Resource{},
				Service: &score.WorkloadService{
					Ports: map[string]score.ServicePort{
						"www": {
							Port: 8080,
						},
					},
				},
			},
			expectedProvisioned: resources.Provisioned{
				Services: map[string]any{
					"test-workload-test-container": compose.ServiceConfig{
						Name:       "test-workload-test-container",
						Image:      ".",
						Entrypoint: []string{"/bin/sh"},
						Environment: compose.MappingWithEquals{
							"MESSAGE": Ref("Hello Environment"),
						},
						Command: []string{"-c", "echo \"Hello Command\""},
						Ports: []compose.ServicePortConfig{
							{
								Target:    8080,
								Published: "8080",
							},
						},
						Volumes: []compose.ServiceVolumeConfig{
							{
								Type:     "bind",
								Source:   "test-context-dir/files/" + HashOfString("Hello File"),
								Target:   "/test/file.txt",
								ReadOnly: true,
							},
						},
					},
				},
				Networks: map[string]any{},
				State:    map[string]any{},
				Files: map[string]any{
					HashOfString("Hello File"): "Hello File",
				},
				Outputs: map[string]any{},
			},
			context: *NewContext("test-context-dir", nil),
		},
		{
			name: "resources in environment",
			workloadSpec: score.Workload{
				ApiVersion: "spec.score.dev/v1b1",
				Containers: map[string]score.Container{
					"test-container": {
						Image:          ".",
						LivenessProbe:  &score.ContainerProbe{},
						ReadinessProbe: &score.ContainerProbe{},
						Resources:      &score.ContainerResources{},
						Variables: map[string]string{
							"BUCKET": "${resources.s3.bucket}",
						},
						Volumes: []score.ContainerVolumesElem{},
					},
				},
				Metadata: map[string]interface{}{
					"name": "test-workload",
				},
				Resources: map[string]score.Resource{
					"s3": {
						Type: "s3",
					},
				},
				Service: &score.WorkloadService{},
			},
			expectedProvisioned: resources.Provisioned{
				Services: map[string]any{
					"test-workload-test-container": compose.ServiceConfig{
						Name:  "test-workload-test-container",
						Image: ".",
						Environment: compose.MappingWithEquals{
							"BUCKET": Ref("my-s3-bucket"),
						},
						Volumes: []compose.ServiceVolumeConfig{},
					},
				},
				Networks: map[string]any{},
				State:    map[string]any{},
				Files:    map[string]any{},
				Outputs:  map[string]any{},
			},
			context: Context{
				ResourceState: ResourceState{
					Workloads: map[string]map[string]string{
						"test-workload": map[string]string{
							"s3": "workloads.test-workload.resources.s3",
						},
					},
					Resources: map[string]*Resource{},
					Provisioned: map[string]*resources.Provisioned{
						"workloads.test-workload.resources.s3": &resources.Provisioned{
							Outputs: map[string]any{
								"bucket": "my-s3-bucket",
							},
						},
					},
				},
				WorkloadState: WorkloadState{
					Specs:       map[string]*score.Workload{},
					Provisioned: map[string]*resources.Provisioned{},
					Build:       map[string]any{},
				},
				appFs:      nil,
				contextDir: "test",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualProvisioned, err := testCase.context.ProvisionWorkload(&testCase.workloadSpec)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedProvisioned, *actualProvisioned)
		})
	}
}

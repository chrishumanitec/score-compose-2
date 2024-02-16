package score

import (
	"testing"

	"github.com/score-spec/score-compose/internal/resources"
	score "github.com/score-spec/score-go/types"
	"github.com/stretchr/testify/assert"
)

func TestUpdateContextResources(t *testing.T) {
	var (
		workloadSpecEmpty = score.Workload{
			ApiVersion: "",
			Containers: map[string]score.Container{},
			Metadata: score.WorkloadMetadata{
				"name": "test-workload-empty",
			},
			Resources: map[string]score.Resource{},
			Service: &score.WorkloadService{
				Ports: map[string]score.ServicePort{},
			},
		}
		workloadSpecOne = score.Workload{
			ApiVersion: "",
			Containers: map[string]score.Container{},
			Metadata: score.WorkloadMetadata{
				"name": "test-workload-one",
			},
			Resources: map[string]score.Resource{
				"test-resource": {
					Class:    nil,
					Metadata: nil,
					Params:   nil,
					Type:     "test-type",
				},
			},
			Service: &score.WorkloadService{
				Ports: map[string]score.ServicePort{},
			},
		}
		resourceStateOne = ResourceState{
			Workloads: map[string]map[string]string{
				workloadSpecOne.Metadata["name"].(string): {
					"test-resource": "test-type::default::test-workload-one.test-resource",
				},
			},
			Resources: map[string]*Resource{
				"test-type::default::test-workload-one.test-resource": &Resource{
					ID:       "test-type::default::test-workload-one.test-resource",
					Current:  true,
					Name:     "test-workload-one.test-resource",
					Resource: workloadSpecOne.Resources["test-resource"],
				},
			},
			Provisioned: map[string]*resources.Provisioned{},
		}
		workloadSpecTwo = score.Workload{
			ApiVersion: "",
			Containers: map[string]score.Container{},
			Metadata: score.WorkloadMetadata{
				"name": "test-workload-two",
			},
			Resources: map[string]score.Resource{
				"test-resource": {
					Class:    nil,
					Metadata: nil,
					Params:   nil,
					Type:     "test-type",
				},
			},
			Service: &score.WorkloadService{
				Ports: map[string]score.ServicePort{},
			},
		}
	)
	testCases := []struct {
		name                  string
		context               *Context
		spec                  score.Workload
		expectedResourceState ResourceState
	}{
		{
			name:    "no resources in score file",
			context: NewContext("", nil),
			spec:    workloadSpecEmpty,
			expectedResourceState: ResourceState{
				Workloads: map[string]map[string]string{
					workloadSpecEmpty.Metadata["name"].(string): {},
				},
				Resources:   map[string]*Resource{},
				Provisioned: map[string]*resources.Provisioned{},
			},
		},
		{
			name:                  "single resource in score file",
			context:               NewContext("", nil),
			spec:                  workloadSpecOne,
			expectedResourceState: resourceStateOne,
		},
		{
			name: "single resource in score file added to existing context",
			context: &Context{
				ResourceState: resourceStateOne,
				WorkloadState: WorkloadState{
					Specs: map[string]*score.Workload{
						workloadSpecOne.Metadata["name"].(string): &workloadSpecOne,
					},
				},
				Version: "",
			},
			spec: workloadSpecTwo,
			expectedResourceState: ResourceState{
				Workloads: map[string]map[string]string{
					workloadSpecOne.Metadata["name"].(string): {
						"test-resource": "test-type::default::test-workload-one.test-resource",
					},
					workloadSpecTwo.Metadata["name"].(string): {
						"test-resource": "test-type::default::test-workload-two.test-resource",
					},
				},
				Resources: map[string]*Resource{
					"test-type::default::test-workload-one.test-resource": {
						ID:       "test-type::default::test-workload-one.test-resource",
						Current:  true,
						Name:     "test-workload-one.test-resource",
						Resource: workloadSpecOne.Resources["test-resource"],
					},
					"test-type::default::test-workload-two.test-resource": {
						ID:       "test-type::default::test-workload-two.test-resource",
						Current:  true,
						Name:     "test-workload-two.test-resource",
						Resource: workloadSpecOne.Resources["test-resource"],
					},
				},
				Provisioned: map[string]*resources.Provisioned{},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.context.Update(&testCase.spec)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResourceState, testCase.context.ResourceState)
		})
	}
}

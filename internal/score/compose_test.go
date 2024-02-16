package score

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeMap(t *testing.T) {
	testCases := []struct {
		name    string
		m1      map[string]any
		m2      map[string]any
		out     map[string]any
		errPath string
	}{
		{
			name: "no overlap",
			m1: map[string]any{
				"one": map[string]any{
					"one.ONE": "oneOne",
				},
			},
			m2: map[string]any{
				"two": map[string]any{
					"two.ONE": "twoOne",
				},
			},
			out: map[string]any{
				"one": map[string]any{
					"one.ONE": "oneOne",
				},
				"two": map[string]any{
					"two.ONE": "twoOne",
				},
			},
		},
		{
			name: "full overlap",
			m1: map[string]any{
				"one": map[string]any{
					"one.ONE": "oneOne",
				},
			},
			m2: map[string]any{
				"one": map[string]any{
					"one.ONE": "oneOne",
				},
			},
			out: map[string]any{
				"one": map[string]any{
					"one.ONE": "oneOne",
				},
			},
		},
		{
			name: "partial overlap",
			m1: map[string]any{
				"one": map[string]any{
					"one.ONE": "oneOne",
				},
			},
			m2: map[string]any{
				"one": map[string]any{
					"one.ONE": "oneOne",
				},
				"two": map[string]any{
					"two.ONE": "twoOne",
				},
			},
			out: map[string]any{
				"one": map[string]any{
					"one.ONE": "oneOne",
				},
				"two": map[string]any{
					"two.ONE": "twoOne",
				},
			},
		},
		{
			name: "partial overlap and fill out",
			m1: map[string]any{
				"one": map[string]any{},
			},
			m2: map[string]any{
				"one": map[string]any{
					"one.ONE": "oneOne",
				},
				"two": map[string]any{
					"two.ONE": "twoOne",
				},
			},
			out: map[string]any{
				"one": map[string]any{
					"one.ONE": "oneOne",
				},
				"two": map[string]any{
					"two.ONE": "twoOne",
				},
			},
		},
		{
			name: "overlap with conflict",
			m1: map[string]any{
				"one": map[string]any{
					"one.ONE": "oneOne",
				},
			},
			m2: map[string]any{
				"one": map[string]any{
					"one.ONE": "else",
				},
			},
			errPath: `test.one["one.ONE"]`,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualOut, err := mergeMap(testCase.m1, testCase.m2, "test")
			if testCase.errPath != "" {
				assert.ErrorContains(t, err, testCase.errPath)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.out, actualOut)
			}
		})
	}
}

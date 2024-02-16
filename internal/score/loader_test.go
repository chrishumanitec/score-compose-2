package score

import (
	"bytes"
	"testing"

	score "github.com/score-spec/score-go/types"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	spec := `
apiVersion: score.sh/v1b1

metadata:
  name: hello-world

containers:
  hello:
    image: "busybox:latest"
    command: ["/bin/sh"]
    args: ["-c", "while true; do echo Hello World!; sleep 5; done"]
    variables:
      DB_NAME: ${resources.db.name}

resources:
  db:
    type: postgres
`
	overrides := []string{
		"containers.hello.image=replaced:container",
	}

	expectedWorkload := score.Workload{
		ApiVersion: "score.sh/v1b1",
		Containers: map[string]score.Container{
			"hello": {
				Args: []string{
					"-c",
					"while true; do echo Hello World!; sleep 5; done",
				},
				Command: []string{"/bin/sh"},
				Image:   "replaced:container",
				Variables: map[string]string{
					"DB_NAME": "${resources.db.name}",
				},
			},
		},
		Metadata: map[string]interface{}{
			"name": "hello-world",
		},
		Resources: map[string]score.Resource{
			"db": {
				Type: "postgres",
			},
		},
	}

	specReader := bytes.NewBufferString(spec)
	workload, err := LoadSpec(specReader, overrides)
	assert.NoError(t, err)
	assert.Equal(t, expectedWorkload, *workload)
}

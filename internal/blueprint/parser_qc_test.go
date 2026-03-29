package blueprint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_MissingRequiredField_ReturnsDescriptiveError(t *testing.T) {
	tmp := t.TempDir()
	missingName := `schema_version: "2"
version: "1.0"
actions:
  - type: shell
    command: "echo"
`
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "template.yaml"), []byte(missingName), 0644))
	_, err := Parse(tmp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestParse_IncompatibleSymphonyVersion_ReturnsError(t *testing.T) {
	tmp := t.TempDir()
	yaml := `schema_version: "2"
name: "x"
version: "1.0"
min_symphony_version: "v99.99.99"
actions:
  - type: shell
    command: "echo"
`
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "template.yaml"), []byte(yaml), 0644))
	_, err := Parse(tmp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Symphony")
}

func TestParse_CircularInheritance_ReturnsError(t *testing.T) {
	fetcher := &MockFetcher{
		t: t,
		mockTemplate: map[string]string{
			"A": `
name: Parent A
version: "1.0"
extends: B
actions:
  - type: shell
    command: A
`,
			"B": `
name: Parent B
version: "1.0"
extends: A
actions:
  - type: shell
    command: B
`,
		},
	}
	child := &Blueprint{Extends: "A"}
	_, err := Resolve(child, fetcher)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular")
}

package blueprint

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockFetcher struct {
	t            *testing.T
	mockTemplate map[string]string
}

func (m *MockFetcher) Fetch(source string) (string, error) {
	yamlContent, ok := m.mockTemplate[source]
	if !ok {
		return "", fmt.Errorf("remote fetch error: source %s tidak ditemukan di mock repo", source)
	}

	dir := m.t.TempDir()
	filepath := filepath.Join(dir, "template.yaml")
	_ = os.WriteFile(filepath, []byte(yamlContent), 0644)
	return dir, nil
}

func TestResolve_MergePromptsOverrides(t *testing.T) {
	fetcher := &MockFetcher{
		t: t,
		mockTemplate: map[string]string{
			"base-template-repo": `
name: Base
version: "1.0"
actions:
  - type: shell
    command: test
prompts:
  - id: PROJ_TYPE
    default: "web"
  - id: REDIS_URL
    default: "localhost"
`,
		},
	}

	child := &Blueprint{
		Name:    "Child App",
		Extends: "base-template-repo",
		Prompts: []Prompt{
			{ID: "PROJ_TYPE", Default: "api"}, // Menimpa Base (Override value)
			{ID: "EXTRA_FLAG", Default: "on"}, // Prompt eksklusif Anak
		},
	}

	resolved, err := Resolve(child, fetcher)
	require.NoError(t, err)

	assert.Equal(t, "Child App", resolved.Name)
	assert.Len(t, resolved.Prompts, 3)

	assert.Equal(t, "PROJ_TYPE", resolved.Prompts[0].ID)
	assert.Equal(t, "api", resolved.Prompts[0].Default)

	assert.Equal(t, "REDIS_URL", resolved.Prompts[1].ID)
	assert.Equal(t, "EXTRA_FLAG", resolved.Prompts[2].ID)
}

func TestResolve_CircularDetection(t *testing.T) {
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

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inheritance berulang (circular inheritance) terdeteksi")
}

func TestResolve_ActionUnion(t *testing.T) {
	fetcher := &MockFetcher{
		t: t,
		mockTemplate: map[string]string{
			"base": `
name: Base
version: "1.0"
actions:
  - type: render
    target: "index.html"
`,
		},
	}

	child := &Blueprint{
		Extends: "base",
		Actions: []Action{
			{Type: "shell", Command: "npm install"},
		},
	}

	resolved, err := Resolve(child, fetcher)
	require.NoError(t, err)

	assert.Len(t, resolved.Actions, 2)
	assert.Equal(t, "index.html", resolved.Actions[0].Target) 
	assert.Equal(t, "shell", resolved.Actions[1].Type)
}

package blueprint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tmpDir := t.TempDir()

	validYaml := `
schema_version: "2"
name: "Test"
version: "1.0"
actions:
  - type: "render"
`
	err := os.WriteFile(filepath.Join(tmpDir, "template.yaml"), []byte(validYaml), 0644)
	assert.NoError(t, err)

	bp, err := Parse(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "Test", bp.Name)

	invalidYaml := `
version: "1.0"
actions:
  - type: "render"
`
	err = os.WriteFile(filepath.Join(tmpDir, "template.yaml"), []byte(invalidYaml), 0644)
	assert.NoError(t, err)

	_, err = Parse(tmpDir)
	assert.ErrorContains(t, err, "name")

	emptyDir := t.TempDir()
	_, err = Parse(emptyDir)
	assert.ErrorContains(t, err, "tidak ditemukan")
}

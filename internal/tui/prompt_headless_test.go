package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/username/symphony/internal/blueprint"
)

// resolvePromptVisibility mengembalikan apakah prompt harus ditampilkan
// berdasarkan jawaban yang sudah dikumpulkan sejauh ini.
// Membungkus panggilan ke blueprint.ResolveVisible.
func resolvePromptVisibility(p blueprint.Prompt, answers map[string]any) (visible bool, err error) {
	prompts, err := blueprint.ResolveVisible([]blueprint.Prompt{p}, answers)
	return len(prompts) > 0, err
}

func TestPromptVisibility_NoDependsOn(t *testing.T) {
	p := blueprint.Prompt{
		ID:       "PROJECT_NAME",
		Question: "Nama proyek:",
		Type:     "input",
	}
	visible, err := resolvePromptVisibility(p, map[string]any{})
	assert.NoError(t, err)
	assert.True(t, visible)
}

func TestPromptVisibility_DependsOnTrue(t *testing.T) {
	p := blueprint.Prompt{
		ID:        "DB_HOST",
		Question:  "Hostname database:",
		Type:      "input",
		DependsOn: "DB_TYPE != 'None'",
	}
	answers := map[string]any{"DB_TYPE": "PostgreSQL"}
	visible, err := resolvePromptVisibility(p, answers)
	assert.NoError(t, err)
	assert.True(t, visible)
}

func TestPromptVisibility_DependsOnFalse(t *testing.T) {
	p := blueprint.Prompt{
		ID:        "DB_HOST",
		Question:  "Hostname database:",
		Type:      "input",
		DependsOn: "DB_TYPE != 'None'",
	}
	answers := map[string]any{"DB_TYPE": "None"}
	visible, err := resolvePromptVisibility(p, answers)
	assert.NoError(t, err)
	assert.False(t, visible, "Ketika DB_TYPE == 'None', DB_HOST tidak boleh ditampilkan")
}

func TestPromptVisibility_ConfirmDepends(t *testing.T) {
	p := blueprint.Prompt{
		ID:        "USE_MIGRATIONS",
		Question:  "Gunakan migrations?",
		Type:      "confirm",
		DependsOn: "DB_TYPE == 'PostgreSQL'",
	}

	// Saat DB_TYPE = PostgreSQL → harus tampil
	answers := map[string]any{"DB_TYPE": "PostgreSQL"}
	visible, err := resolvePromptVisibility(p, answers)
	assert.NoError(t, err)
	assert.True(t, visible)

	// Saat DB_TYPE = MongoDB → harus di-skip
	answers["DB_TYPE"] = "MongoDB"
	visible, err = resolvePromptVisibility(p, answers)
	assert.NoError(t, err)
	assert.False(t, visible)
}

func TestPromptVisibility_BoolDepends(t *testing.T) {
	p := blueprint.Prompt{
		ID:        "REDIS_MAX_CONN",
		Question:  "Max Redis connections:",
		Type:      "input",
		DependsOn: "USE_REDIS == true",
	}

	answers := map[string]any{"USE_REDIS": true}
	visible, err := resolvePromptVisibility(p, answers)
	assert.NoError(t, err)
	assert.True(t, visible)

	answers["USE_REDIS"] = false
	visible, err = resolvePromptVisibility(p, answers)
	assert.NoError(t, err)
	assert.False(t, visible)
}

func TestDependencyChain(t *testing.T) {
	// Simulasi alur lengkap symphony-phase-02-tui.md contoh:
	// Prompt 1: DB_TYPE → user jawab "None"
	// Prompt 2: DB_HOST (depends_on: "DB_TYPE != 'None'") → SKIP
	// Prompt 3: USE_MIGRATIONS (depends_on: "DB_TYPE == 'PostgreSQL'") → SKIP
	// Prompt 4: USE_REDIS → tampilkan (tidak ada depends_on)

	answers := map[string]any{}

	// Step 1: DB_TYPE -- tidak ada depends_on, selalu tampil
	p1 := blueprint.Prompt{ID: "DB_TYPE", DependsOn: ""}
	v1, _ := resolvePromptVisibility(p1, answers)
	assert.True(t, v1)
	answers["DB_TYPE"] = "None"

	// Step 2: DB_HOST -- depends_on: "DB_TYPE != 'None'"
	p2 := blueprint.Prompt{ID: "DB_HOST", DependsOn: "DB_TYPE != 'None'"}
	v2, _ := resolvePromptVisibility(p2, answers)
	assert.False(t, v2, "DB_HOST harus di-skip saat DB_TYPE == 'None'")

	// Step 3: USE_MIGRATIONS -- depends_on: "DB_TYPE == 'PostgreSQL'"
	p3 := blueprint.Prompt{ID: "USE_MIGRATIONS", DependsOn: "DB_TYPE == 'PostgreSQL'"}
	v3, _ := resolvePromptVisibility(p3, answers)
	assert.False(t, v3, "USE_MIGRATIONS harus di-skip saat DB_TYPE == 'None'")

	// Step 4: USE_REDIS -- tidak ada depends_on, selalu tampil
	p4 := blueprint.Prompt{ID: "USE_REDIS", DependsOn: ""}
	v4, _ := resolvePromptVisibility(p4, answers)
	assert.True(t, v4, "USE_REDIS harus tampil")
}

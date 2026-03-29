package blueprint

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestValidateDependencyGraph_NoCycles(t *testing.T) {
	prompts := []Prompt{
		{ID: "A"}, // Base
		{ID: "B", DependsOn: "A == 'Yes'"},
		{ID: "C", DependsOn: "B == 'Yes' && A == 'Yes'"}, // Linear diamond dependency
	}
	err := ValidateDependencyGraph(prompts)
	assert.NoError(t, err)
}

func TestValidateDependencyGraph_UnknownPrompt(t *testing.T) {
	prompts := []Prompt{
		{ID: "A"},
		{ID: "B", DependsOn: "UNKNOWN == 'Yes'"}, // Unknown depend
	}
	err := ValidateDependencyGraph(prompts)
	// Kita masih butuh prompt ID tersebut dari evaluate gval, tapi jika regex gagal menemukannya, ia hanya dianggap bukan dependency ke prompt lain di scope ini.
	// Oh tunggu, jika tidak ditemukan di regex, adjacency list-nya kosong, jadi valid.
	// Kita tidak mau memvalidasi syntax gval di sini (itu tugas pkg/expr gval evaluasi).
	// Ini hanya memvalidasi topological graf antar variabel prompt yang kita punya.
	assert.NoError(t, err) // Ini wajar dan harus dibiarkan karena bisa jadi itu variabel context global (seperti OS, dll).
}

func TestValidateDependencyGraph_CircularDirect(t *testing.T) {
	prompts := []Prompt{
		{ID: "A", DependsOn: "B == 'Yes'"},
		{ID: "B", DependsOn: "A == 'Yes'"},
	}
	err := ValidateDependencyGraph(prompts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency terdeteksi")
}

func TestValidateDependencyGraph_CircularIndirect(t *testing.T) {
	prompts := []Prompt{
		{ID: "A", DependsOn: "C == 'Yes'"},
		{ID: "B", DependsOn: "A == 'Yes'"},
		{ID: "C", DependsOn: "B == 'Yes'"},
	}
	err := ValidateDependencyGraph(prompts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency terdeteksi")
}

func TestResolveVisible_ResolvesProperly(t *testing.T) {
	prompts := []Prompt{
		{ID: "A"},
		{ID: "B", DependsOn: "A == true"}, // B visible if A==true
		{ID: "C", DependsOn: "A == false"}, // C visible if A==false
	}

	visible, err := ResolveVisible(prompts, map[string]any{"A": true})
	assert.NoError(t, err)
	assert.Len(t, visible, 2)
	assert.Equal(t, "A", visible[0].ID)
	assert.Equal(t, "B", visible[1].ID)

	visible2, err2 := ResolveVisible(prompts, map[string]any{"A": false})
	assert.NoError(t, err2)
	assert.Len(t, visible2, 2)
	assert.Equal(t, "A", visible2[0].ID)
	assert.Equal(t, "C", visible2[1].ID)
}

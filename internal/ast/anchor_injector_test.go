package ast

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/username/symphony/internal/blueprint"
)

func TestAnchorInjector_BasicStrategy(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	content := "Line 1\n# SECTION: Routes\nLine 2"
	require.NoError(t, os.WriteFile(file, []byte(content), 0644))

	injector := AnchorInjector{}

	// Buka blok injeksi AFTER 
	act1 := blueprint.Action{
		Anchor:   "# SECTION: Routes",
		Strategy: "after-anchor",
		Content:  "Line 1.5",
	}
	require.NoError(t, injector.Inject(file, act1))

	res, _ := os.ReadFile(file)
	actual := string(res)
	assert.Contains(t, actual, "SECTION: Routes\nLine 1.5\nLine 2")
}

func TestAnchorInjector_NotFoundSafety(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	content := "Line 1\n# SECTION: Unmatched\nLine 2"
	require.NoError(t, os.WriteFile(file, []byte(content), 0644))

	injector := AnchorInjector{}
	act := blueprint.Action{
		Anchor:   "# SECTION: Match Me",
		Strategy: "after-anchor",
		Content:  "Injection!",
	}
	
	err := injector.Inject(file, act)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "string anchor '# SECTION: Match Me' tidak ditemukan")

	// Pastikan isi file tak tersentuh
	res, _ := os.ReadFile(file)
	assert.Equal(t, content, string(res))
	
	// Pastikan backup log terhapus
	_, statErr := os.Stat(file + ".symphony-bak")
	assert.True(t, os.IsNotExist(statErr))
}

func TestGoInjector_SyntaxValidFormatter(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "main.go")
	content := `package main

import "fmt"

func main() {
	// @ROUTES
	fmt.Println("Start")
}
`
	require.NoError(t, os.WriteFile(file, []byte(content), 0644))
	inj := &GoInjector{}

	act := blueprint.Action{
		Anchor:   "// @ROUTES",
		Strategy: "after-anchor",
		Content:  "fmt.Println(\"Injected Route\")",
	}

	err := inj.Inject(file, act)
	assert.NoError(t, err)

	res, _ := os.ReadFile(file)
	actual := string(res)
	// Kita periksa injeksi berjalan dengan melihat string aslinya.
	assert.Contains(t, actual, "fmt.Println(\"Start\")")
	assert.Contains(t, actual, "fmt.Println(\"Injected Route\")")
}

func TestGoInjector_SyntaxInvalid(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "main.go")
	content := `package main

func main() {
	// @ANCHOR
}`
	require.NoError(t, os.WriteFile(file, []byte(content), 0644))

	inj := &GoInjector{}

	act := blueprint.Action{
		Anchor:   "// @ANCHOR",
		Strategy: "after-anchor",
		Content:  "broken_syntax( ) { !!", // Merusak sintaks
	}

	err := inj.Inject(file, act)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sintak salah")

	// Berkas berhasil direstorasi setelah merusak dan tak mengubah apa-apa
	res, _ := os.ReadFile(file)
	assert.Equal(t, content, string(res))
}

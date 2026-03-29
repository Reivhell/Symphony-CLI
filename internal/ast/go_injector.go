package ast

import (
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/username/symphony/internal/blueprint"
)

// GoInjector bertindak selaku penjaga wrapper injeksi teks anchor
// khusus file berbahasa Go dengan garansi `gofmt` pasca update.
type GoInjector struct {
	anchor AnchorInjector
}

func (g *GoInjector) CanHandle(filePath string) bool {
	return strings.HasSuffix(filePath, ".go")
}

func (g *GoInjector) Inject(targetPath string, action blueprint.Action) error {
	backupPath := targetPath + ".symphony-bak-go"
	if err := copyFile(targetPath, backupPath); err != nil {
		return fmt.Errorf("gagal membuat backup golang injeksi: %w", err)
	}
	
	success := false
	defer func() {
		if !success { // Pulihkan karena AST Golang tidak cocok atau syntax error!
			_ = copyFile(backupPath, targetPath)
		}
		_ = os.Remove(backupPath)
	}()

	// Titipkan sisip anchor teks biasa
	if err := g.anchor.Inject(targetPath, action); err != nil {
		return err // success masih false, file aman di restore
	}

	// Langkah 2: Evaluasi sintaks pasca file dimasuki kode injection (AST Go Validation)
	newContent, err := os.ReadFile(targetPath)
	if err != nil {
		return fmt.Errorf("gagal mereview injeksi .go: %w", err)
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", newContent, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("gagal compile go injeksi (sintak salah): %w", err)
	}

	// Format kodenya yang benar layaknya 'gofmt'
	formatted, err := gofmt(fset, f)
	if err != nil {
		return fmt.Errorf("injeksi gagal pada proses gofmt internal: %w", err)
	}

	if err := os.WriteFile(targetPath, formatted, 0644); err != nil {
		return fmt.Errorf("gagal meresave injeksi berformat standar: %w", err)
	}

	success = true
	return nil
}

func gofmt(fset *token.FileSet, f interface{}) ([]byte, error) {
	var buf strings.Builder
	if err := format.Node(&buf, fset, f); err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}

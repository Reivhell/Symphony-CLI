package ast

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/username/symphony/internal/blueprint"
)

// AnchorInjector adalah modul dasar untuk penambahan (injeksi)
// yang mengandalkan kemiripan teks statis string atau identifikasi komentar
// di dalam file kode sumber apa pun sebagai penanda posisi (anchor).
type AnchorInjector struct{}

// CanHandle selalu bernilai benar.
// Pengganti string dinamis ini bersifat agnostik dan mengenali text file apapun.
func (a *AnchorInjector) CanHandle(filePath string) bool {
	return true
}

func (a *AnchorInjector) Inject(targetPath string, action blueprint.Action) error {
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("injeksi gagal: file target %s tidak ditemukan", targetPath)
	}

	backupPath := targetPath + ".symphony-bak"
	if err := copyFile(targetPath, backupPath); err != nil {
		return fmt.Errorf("gagal membuat backup file target: %w", err)
	}

	// Tangani Pemulihan jika gagal
	success := false
	defer func() {
		if !success {
			_ = copyFile(backupPath, targetPath)
		}
		// Hapus backup di akhir operasi entah ia sukses/gagal me-replace ulang
		_ = os.Remove(backupPath)
	}()

	fileContent, err := os.ReadFile(targetPath)
	if err != nil {
		return fmt.Errorf("gagal membaca file: %w", err)
	}

	lines := strings.Split(string(fileContent), "\n")
	var result []string
	injected := false

	anchor := strings.TrimSpace(action.Anchor)

	for _, line := range lines {
		if !injected && strings.Contains(line, anchor) {
			indent := extractIndent(line)
			contentLines := indentBlock(action.Content, indent)

			switch action.Strategy {
			case "after-anchor":
				result = append(result, line)
				result = append(result, contentLines...)
			case "before-anchor":
				result = append(result, contentLines...)
				result = append(result, line)
			case "replace-anchor":
				result = append(result, contentLines...)
			default:
				return fmt.Errorf("strategi anchor '%s' tidak didukung", action.Strategy)
			}
			injected = true
		} else {
			result = append(result, line)
		}
	}

	if !injected {
		return fmt.Errorf("injeksi gagal: string anchor '%s' tidak ditemukan di dokumen", action.Anchor)
	}

	newContent := strings.Join(result, "\n")
	if err := os.WriteFile(targetPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("gagal menulis ulang file yang disisipkan: %w", err)
	}

	success = true
	return nil
}

func extractIndent(line string) string {
	for i, c := range line {
		if c != ' ' && c != '\t' {
			return line[:i]
		}
	}
	return ""
}

func indentBlock(content, indent string) []string {
	var indented []string
	reader := bufio.NewReader(bytes.NewBufferString(content))
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		
		if line == "" && err == io.EOF {
			break
		}
		
		if line != "" { // cegah indentasi statis pada baris kosong
			indented = append(indented, indent+line)
		} else {
			indented = append(indented, "")
		}

		if err == io.EOF {
			break
		}
	}
	return indented
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

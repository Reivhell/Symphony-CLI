package blueprint

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/username/symphony/internal/version"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"
)

// Parse membaca file template.yaml dari path yang diberikan,
// memvalidasi strukturnya, dan mengembalikan Blueprint yang sudah di-parse.
func Parse(templateDir string) (*Blueprint, error) {
	yamlPath := filepath.Join(templateDir, "template.yaml")

	data, err := os.ReadFile(yamlPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("file template.yaml tidak ditemukan di direktori: %s", templateDir)
		}
		return nil, fmt.Errorf("gagal membaca template.yaml: %w", err)
	}

	var bp Blueprint
	if err := yaml.Unmarshal(data, &bp); err != nil {
		return nil, fmt.Errorf("gagal mem-parse YAML: %w", err)
	}

	if strings.TrimSpace(bp.Name) == "" {
		return nil, errors.New("validasi gagal: field 'name' tidak boleh kosong")
	}
	if strings.TrimSpace(bp.Version) == "" {
		return nil, errors.New("validasi gagal: field 'version' tidak boleh kosong")
	}
	if len(bp.Actions) == 0 {
		return nil, errors.New("validasi gagal: minimal harus ada satu 'action'")
	}

	if bp.SchemaVersion != "" && bp.SchemaVersion != "2" {
		return nil, fmt.Errorf("validasi gagal: schema_version tidak didukung, diharapkan '2' tapi mendapat '%s'", bp.SchemaVersion)
	}

	if bp.MinSymphonyVersion != "" {
		req := canonicalSemver(bp.MinSymphonyVersion)
		current := canonicalSemver(version.Version)
		if semver.Compare(req, current) > 0 {
			return nil, fmt.Errorf("versi Symphony CLI tidak kompatibel. Template membutuhkan versi %s (saat ini %s)", req, current)
		}
	}

	if err := ValidateDependencyGraph(bp.Prompts); err != nil {
		return nil, fmt.Errorf("validasi prompt gagal: %w", err)
	}

	return &bp, nil
}

// canonicalSemver returns a semver string comparable by golang.org/x/mod/semver.
// Non-semver values (e.g. "dev") fall back to v0.1.0 for comparison.
func canonicalSemver(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "v0.1.0"
	}
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	if semver.IsValid(v) {
		return semver.Canonical(v)
	}
	return "v0.1.0"
}

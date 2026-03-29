package blueprint

import (
	"fmt"
)

// Fetcher adalah antarmuka untuk mengambil source blueprint secara dinamis
// Di-inject dalam sistem Resolving warisan file
type Fetcher interface {
	Fetch(source string) (localPath string, err error)
}

// Resolve menangani rekursif template inheritance (extends).
// Ia menanggulangi circular references dan melakukan penggabungan (merge) proporsi template.
func Resolve(child *Blueprint, fetcher Fetcher) (*Blueprint, error) {
	return resolveRecursive(child, fetcher, map[string]bool{})
}

func resolveRecursive(current *Blueprint, fetcher Fetcher, visited map[string]bool) (*Blueprint, error) {
	if current.Extends == "" {
		return current, nil
	}

	if visited[current.Extends] {
		return nil, fmt.Errorf("inheritance berulang (circular inheritance) terdeteksi pada referensi: %s", current.Extends)
	}
	visited[current.Extends] = true

	basePath, err := fetcher.Fetch(current.Extends)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil template induk %s: %w", current.Extends, err)
	}

	parentBlueprint, err := Parse(basePath)
	if err != nil {
		return nil, fmt.Errorf("gagal mem-parsing template induk dari %s: %w", current.Extends, err)
	}

	// Tangani kasus Multilevel inheritance (kakek -> bapak -> anak)
	resolvedParent, err := resolveRecursive(parentBlueprint, fetcher, visited)
	if err != nil {
		return nil, err
	}

	return mergeBlueprints(resolvedParent, current), nil
}

// mergeBlueprints menyatukan template anak dan basis menurut spesifikasi.
func mergeBlueprints(base, child *Blueprint) *Blueprint {
	merged := &Blueprint{
		// 1. Metadata selalu memakai punya anak
		SchemaVersion:      child.SchemaVersion,
		Name:               child.Name,
		Version:            child.Version,
		Author:             child.Author,
		Description:        child.Description,
		MinSymphonyVersion: child.MinSymphonyVersion,
		Tags:               child.Tags,
		
		// 2. Completion Message: utamakan anak, fallback ibu
		CompletionMessage: child.CompletionMessage,
	}

	if merged.CompletionMessage == "" {
		merged.CompletionMessage = base.CompletionMessage
	}

	// 3. Merging Validations: Union semuanya (Base + Child) bebas konflik
	merged.Validations = append(base.Validations, child.Validations...)

	// 4. Prompts: Base duluan; kemudian Child dapat meng-override yang memiliki ID sama.
	merged.Prompts = mergePrompts(base.Prompts, child.Prompts)

	// 5. Actions: Append sequentially. Basis lalu Child. (Anak tak bisa menghapus Basis)
	merged.Actions = append(base.Actions, child.Actions...)

	// Opsional: merge plugin tambahan dsb
	merged.Plugins = append(base.Plugins, child.Plugins...)

	return merged
}

func mergePrompts(base, child []Prompt) []Prompt {
	var result []Prompt
	childMap := make(map[string]Prompt)

	for _, p := range child {
		childMap[p.ID] = p
	}

	// Tuang base dulu, jika ada anak yang punya id sama, ganti valuenya dari anak
	for _, bp := range base {
		if cp, ok := childMap[bp.ID]; ok {
			result = append(result, cp)
			delete(childMap, bp.ID)
		} else {
			result = append(result, bp)
		}
	}

	// Anak bisa menambahkan prompt independen murni baru di posisi bawah
	for _, p := range child {
		if _, exists := childMap[p.ID]; exists {
			result = append(result, p)
		}
	}

	return result
}

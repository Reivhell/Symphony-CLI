package blueprint

import (
	"fmt"
	"regexp"

	"github.com/username/symphony/pkg/expr"
)

// ResolveVisible mengevaluasi mana saja prompt yang harus ditampilkan (visible)
// berdasarkan jawaban (context) yang sudah terkumpul sejauh ini.
func ResolveVisible(prompts []Prompt, collectedAnswers map[string]any) ([]Prompt, error) {
	var visible []Prompt
	for _, p := range prompts {
		// Jika tidak ada depends_on, selalu visible
		if p.DependsOn == "" {
			visible = append(visible, p)
			continue
		}

		res, err := expr.Evaluate(p.DependsOn, collectedAnswers)
		if err != nil {
			return nil, fmt.Errorf("evaluasi depends_on untuk '%s' gagal: %w", p.ID, err)
		}

		if res {
			visible = append(visible, p)
		}
	}
	return visible, nil
}

// ValidateDependencyGraph memastikan rantai depends_on antarprompt tidak mengandung
// siklus dependensi (circular dependency) dan tidak bergantung pada variabel prompt fiktif.
func ValidateDependencyGraph(prompts []Prompt) error {
	// 1. Simpan semua prompt ID yang valid
	validIDs := map[string]bool{}
	for _, p := range prompts {
		validIDs[p.ID] = true
	}

	// 2. Bangun adjacency list (graph berarah): adj[A] = [B, C] berarti A depends on B dan C
	adj := map[string][]string{}
	for _, p := range prompts {
		if p.DependsOn == "" {
			continue
		}
		
		for id := range validIDs {
			// Cari substring kata penuh (word boundaries) untuk memastikan tidak matching variabel parsial (misal match DB_URL pada DB_URL_HOST)
			pattern := `\b` + regexp.QuoteMeta(id) + `\b`
			matched, err := regexp.MatchString(pattern, p.DependsOn)
			if err == nil && matched {
				adj[p.ID] = append(adj[p.ID], id)
			}
		}
	}

	// 3. DFS untuk mencari siklus (3-color marker: 0=unvisited, 1=visiting, 2=visited)
	state := map[string]int{}
	
	var dfs func(node string) error
	dfs = func(node string) error {
		state[node] = 1 // sedang di-kunjungi dalam branch chain saat ini
		
		for _, neighbor := range adj[node] {
			if !validIDs[neighbor] {
				return fmt.Errorf("prompt '%s' depends on undefined prompt: '%s'", node, neighbor)
			}
			if state[neighbor] == 1 {
				// Siklus terdeteksi! (Mengunjungi kembali node yang masih 'visiting' di jalur ini)
				return fmt.Errorf("circular dependency terdeteksi: ketergantungan putaran antara prompt '%s' dan '%s'", node, neighbor)
			}
			if state[neighbor] == 0 {
				if err := dfs(neighbor); err != nil {
					return err
				}
			}
		}
		state[node] = 2 // beres di-kunjungi secara rekursif
		return nil
	}

	for _, p := range prompts {
		if state[p.ID] == 0 {
			if err := dfs(p.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

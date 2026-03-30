package tui

import (
	"fmt"

	"github.com/Reivhell/symphony/internal/blueprint"
)

// RunPrompts menjalankan sesi interaktif lengkap untuk semua prompt dalam blueprint.
func RunPrompts(bp *blueprint.Blueprint) (map[string]any, error) {
	answers := make(map[string]any)

	for _, prompt := range bp.Prompts {
		visible, err := blueprint.ResolveVisible([]blueprint.Prompt{prompt}, answers)
		if err != nil {
			return nil, fmt.Errorf("evaluasi dependency gagal pada prompt %s: %w", prompt.ID, err)
		}
		
		if len(visible) == 0 {
			answers[prompt.ID] = prompt.Default
			continue
		}

		switch prompt.Type {
		case "input":
			ans, err := RunInput(prompt)
			if err != nil {
				return nil, err
			}
			answers[prompt.ID] = ans
		case "select":
			ans, err := RunSelect(prompt)
			if err != nil {
				return nil, err
			}
			answers[prompt.ID] = ans
		case "multiselect":
			ans, err := RunMultiSelect(prompt)
			if err != nil {
				return nil, err
			}
			answers[prompt.ID] = ans
		case "confirm":
			ans, err := RunConfirm(prompt)
			if err != nil {
				return nil, err
			}
			answers[prompt.ID] = ans
		default:
			return nil, fmt.Errorf("tipe prompt tidak dikenal: %s", prompt.Type)
		}
	}

	return answers, nil
}

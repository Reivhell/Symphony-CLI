package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Reivhell/symphony/internal/blueprint"
)

type multiselectModel struct {
	q        blueprint.Prompt
	cursor   int
	selected map[int]struct{}
	done     bool
	err      error
}

func initialMultiSelectModel(q blueprint.Prompt) multiselectModel {
	selected := make(map[int]struct{})
	if q.Default != nil {
		if defSlice, ok := q.Default.([]interface{}); ok {
			for i, opt := range q.Options {
				for _, def := range defSlice {
					if opt == fmt.Sprintf("%v", def) {
						selected[i] = struct{}{}
					}
				}
			}
		}
	}
	return multiselectModel{
		q:        q,
		selected: selected,
	}
}

func (m multiselectModel) Init() tea.Cmd {
	return nil
}

func (m multiselectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.err = ErrUserCancelled
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.q.Options) - 1
			}
		case "down", "j":
			if m.cursor < len(m.q.Options)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		case "enter":
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m multiselectModel) View() string {
	if m.done {
		var selectedOpts []string
		for i, opt := range m.q.Options {
			if _, ok := m.selected[i]; ok {
				selectedOpts = append(selectedOpts, opt)
			}
		}
		return fmt.Sprintf("  %s %s\n  %s %s\n", StyleBrand.Render("?"), m.q.Question, StyleBrand.Render(IconSuccess), StyleMuted.Render(strings.Join(selectedOpts, ", ")))
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("  %s %s %s\n", StyleBrand.Render("?"), m.q.Question, StyleMuted.Render("(Space untuk toggle)")))

	for i, choice := range m.q.Options {
		cursor := "  "
		if m.cursor == i {
			cursor = StyleBrand.Render(IconArrow)
		}

		_, checked := m.selected[i]
		check := "☐"
		if checked {
			check = "☑"
		}

		if m.cursor == i {
			b.WriteString(fmt.Sprintf("  %s %s %s\n", cursor, StyleHighlight.Render(check), StyleHighlight.Render(choice)))
		} else {
			b.WriteString(fmt.Sprintf("  %s %s %s\n", cursor, check, choice))
		}
	}
	return b.String()
}

// RunMultiSelect menjalankan prompt multiselect
func RunMultiSelect(q blueprint.Prompt) ([]string, error) {
	p := tea.NewProgram(initialMultiSelectModel(q))
	m, err := p.Run()
	if err != nil {
		return nil, err
	}
	model := m.(multiselectModel)
	if model.err != nil {
		return nil, model.err
	}

	var result []string
	for i, opt := range model.q.Options {
		if _, ok := model.selected[i]; ok {
			result = append(result, opt)
		}
	}
	return result, nil
}

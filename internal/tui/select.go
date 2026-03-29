package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/username/symphony/internal/blueprint"
)

type selectModel struct {
	q        blueprint.Prompt
	cursor   int
	selected string
	done     bool
	err      error
}

func initialSelectModel(q blueprint.Prompt) selectModel {
	cursor := 0
	if q.Default != nil {
		defStr := fmt.Sprintf("%v", q.Default)
		for i, opt := range q.Options {
			if opt == defStr {
				cursor = i
				break
			}
		}
	}
	return selectModel{
		q:      q,
		cursor: cursor,
	}
}

func (m selectModel) Init() tea.Cmd {
	return nil
}

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "enter":
			if len(m.q.Options) > 0 {
				m.selected = m.q.Options[m.cursor]
			}
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m selectModel) View() string {
	if m.done {
		return fmt.Sprintf("  %s %s\n  %s %s\n", StyleBrand.Render("?"), m.q.Question, StyleBrand.Render(IconSuccess), StyleMuted.Render(m.selected))
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("  %s %s\n", StyleBrand.Render("?"), m.q.Question))

	for i, choice := range m.q.Options {
		cursor := "  "
		if m.cursor == i {
			cursor = StyleBrand.Render(IconArrow)
		}

		check := "○"
		if m.cursor == i {
			check = "●"
			b.WriteString(fmt.Sprintf("  %s %s %s\n", cursor, StyleHighlight.Render(check), StyleHighlight.Render(choice)))
		} else {
			b.WriteString(fmt.Sprintf("  %s %s %s\n", cursor, check, choice))
		}
	}
	return b.String()
}

// RunSelect menjalankan prompt select
func RunSelect(q blueprint.Prompt) (string, error) {
	p := tea.NewProgram(initialSelectModel(q))
	m, err := p.Run()
	if err != nil {
		return "", err
	}
	model := m.(selectModel)
	if model.err != nil {
		return "", model.err
	}
	return model.selected, nil
}

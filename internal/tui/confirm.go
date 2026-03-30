package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Reivhell/symphony/internal/blueprint"
)

type confirmModel struct {
	q        blueprint.Prompt
	done     bool
	err      error
	result   bool
	strTrue  string
	strFalse string
}

func initialConfirmModel(q blueprint.Prompt) confirmModel {
	def := false
	if q.Default != nil {
		if b, ok := q.Default.(bool); ok {
			def = b
		}
	}

	strTrue, strFalse := "y", "N"
	if def {
		strTrue, strFalse = "Y", "n"
	}

	return confirmModel{
		q:        q,
		result:   def,
		strTrue:  strTrue,
		strFalse: strFalse,
	}
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.err = ErrUserCancelled
			return m, tea.Quit
		case "y", "Y":
			m.result = true
			m.done = true
			return m, tea.Quit
		case "n", "N":
			m.result = false
			m.done = true
			return m, tea.Quit
		case "enter":
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	if m.done {
		resStr := "No"
		if m.result {
			resStr = "Yes"
		}
		return fmt.Sprintf("  %s %s\n  %s %s\n", StyleBrand.Render("?"), m.q.Question, StyleBrand.Render(IconSuccess), StyleMuted.Render(resStr))
	}

	return fmt.Sprintf("  %s %s (%s/%s)\n", StyleBrand.Render("?"), m.q.Question, m.strTrue, m.strFalse)
}

// RunConfirm menjalankan prompt confirm
func RunConfirm(q blueprint.Prompt) (bool, error) {
	p := tea.NewProgram(initialConfirmModel(q))
	m, err := p.Run()
	if err != nil {
		return false, err
	}
	model := m.(confirmModel)
	if model.err != nil {
		return false, model.err
	}
	return model.result, nil
}

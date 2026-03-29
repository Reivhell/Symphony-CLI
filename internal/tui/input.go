package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/username/symphony/internal/blueprint"
)

type inputModel struct {
	prompt   textinput.Model
	q        blueprint.Prompt
	done     bool
	err      error
	finalVal string
}

func initialInputModel(q blueprint.Prompt) inputModel {
	ti := textinput.New()
	if q.Default != nil {
		ti.Placeholder = fmt.Sprintf("%v", q.Default)
	}
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return inputModel{
		prompt: ti,
		q:      q,
	}
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.err = ErrUserCancelled
			return m, tea.Quit
		case tea.KeyEnter:
			m.done = true
			m.finalVal = m.prompt.Value()
			if m.finalVal == "" && m.q.Default != nil {
				m.finalVal = fmt.Sprintf("%v", m.q.Default)
			}
			return m, tea.Quit
		}
	}

	m.prompt, cmd = m.prompt.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	if m.done {
		return fmt.Sprintf("  %s %s\n  %s %s\n", StyleBrand.Render("?"), m.q.Question, StyleBrand.Render(IconSuccess), StyleMuted.Render(m.finalVal))
	}

	return fmt.Sprintf("  %s %s\n  %s %s\n", StyleBrand.Render("?"), m.q.Question, StyleBrand.Render(IconArrow), m.prompt.View())
}

// RunInput menjalankan prompt input
func RunInput(q blueprint.Prompt) (string, error) {
	p := tea.NewProgram(initialInputModel(q))
	m, err := p.Run()
	if err != nil {
		return "", err
	}
	model := m.(inputModel)
	if model.err != nil {
		return "", model.err
	}
	return model.finalVal, nil
}

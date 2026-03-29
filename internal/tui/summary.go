package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/username/symphony/internal/engine"
)

type summaryModel struct {
	actions  []engine.ResolvedAction
	done     bool
	err      error
	result   bool
}

func initialSummaryModel(actions []engine.ResolvedAction) summaryModel {
	return summaryModel{
		actions: actions,
	}
}

func (m summaryModel) Init() tea.Cmd {
	return nil
}

func (m summaryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.result = true
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m summaryModel) View() string {
	if m.done {
		resStr := "No"
		if m.result {
			resStr = "Yes"
		}
		return fmt.Sprintf("  %s %s\n  %s %s\n\n", StyleBrand.Render("?"), "Lanjutkan?", StyleBrand.Render(IconSuccess), StyleMuted.Render(resStr))
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s %s\n", StyleBrand.Render(IconDiamond), StyleHeader.Render("Review — File yang akan dibuat:")))
	b.WriteString(Divider(54) + "\n\n")

	createCnt := 0
	skipCnt := 0
	modCnt := 0

	for _, a := range m.actions {
		if !a.ShouldRun {
			src := a.Original.Source
			if src == "" {
				src = a.Original.Target
			}
			reason := ""
			if a.SkipReason != "" {
				reason = StyleMuted.Render("← " + a.SkipReason)
			}
			b.WriteString(fmt.Sprintf("  %-25s %-40s %s\n",
				StyleActionSkip.Render("[SKIP]"),
				StyleActionSkip.Render(src),
				reason,
			))
			skipCnt++
			continue
		}

		switch a.Original.Type {
		case "shell":
			b.WriteString(fmt.Sprintf("  %-25s %s\n", StyleActionModify.Render("[HOOK]"), a.Original.Command))
		case "ast-inject":
			b.WriteString(fmt.Sprintf("  %-25s %s\n", StyleActionModify.Render("[AST-INJ]"), a.Original.Target))
			modCnt++
		default:
			b.WriteString(fmt.Sprintf("  %-25s %s\n", StyleActionCreate.Render("[CREATE]"), a.TargetPath))
			createCnt++
		}
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Total: %d file akan dibuat, %d file dimodifikasi, %d file di-skip\n\n", createCnt, modCnt, skipCnt))
	b.WriteString(fmt.Sprintf("  %s %s (Y/n)\n", StyleBrand.Render(IconArrow), "Lanjutkan?"))

	return b.String()
}

// RenderSummary merender summary dari aksi dan menunggu konfirmasi.
func RenderSummary(actions []engine.ResolvedAction) (bool, error) {
	p := tea.NewProgram(initialSummaryModel(actions))
	m, err := p.Run()
	if err != nil {
		return false, err
	}
	model := m.(summaryModel)
	if model.err != nil {
		return false, model.err
	}
	return model.result, nil
}

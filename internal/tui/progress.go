package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Msg types untuk diurus oleh progress bar
type progressMsg struct {
	Total     int
	Current   int
	File      string
	Action    string
	Completed bool
}

type hookStartMsg struct {
	Command string
}

type hookOutputMsg struct {
	Line string
}

type progressEndedMsg struct{}

type progressModel struct {
	prog       progress.Model
	spin       spinner.Model
	files      []string
	total      int
	current    int
	progWidth  int
	ch         <-chan interface{}
	done       bool
	hookLine   string
	hookCmd    string
	processing bool
}

func initialProgressModel(total int, updates <-chan interface{}) progressModel {
	p := progress.New(progress.WithDefaultGradient())
	p.Width = 50

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = StyleBrand

	return progressModel{
		prog:      p,
		spin:      s,
		total:     total,
		ch:        updates,
		progWidth: 50,
	}
}

func waitForActivity(sub <-chan interface{}) tea.Cmd {
	return func() tea.Msg {
		v, ok := <-sub
		if !ok {
			return progressEndedMsg{}
		}
		return v
	}
}

func (m progressModel) Init() tea.Cmd {
	return tea.Batch(
		m.spin.Tick,
		waitForActivity(m.ch),
	)
}

func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.progWidth = msg.Width - 14
		if m.progWidth > 50 {
			m.progWidth = 50
		}
		m.prog.Width = m.progWidth

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case progressMsg:
		if msg.Completed {
			m.current = msg.Total
			m.done = true
			return m, tea.Quit
		}
		m.current = msg.Current
		if msg.File != "" {
			m.files = append(m.files, msg.File)
		}
		// limit to last 5
		if len(m.files) > 5 {
			m.files = m.files[len(m.files)-5:]
		}
		m.processing = true
		return m, waitForActivity(m.ch)

	case hookStartMsg:
		m.hookCmd = msg.Command
		m.processing = true
		return m, waitForActivity(m.ch)

	case hookOutputMsg:
		m.hookLine = msg.Line
		return m, waitForActivity(m.ch)

	case progressEndedMsg:
		m.done = true
		return m, tea.Quit
	}

	return m, nil
}

func (m progressModel) View() string {
	b := strings.Builder{}

	b.WriteString(fmt.Sprintf("%s %s\n", StyleBrand.Render(IconDiamond), StyleHeader.Render("Generating...")))
    b.WriteString(Divider(54) + "\n\n")

	for i, f := range m.files {
		if i == len(m.files)-1 && m.processing && !m.done {
			b.WriteString(fmt.Sprintf("  %s %s\n", m.spin.View(), f))
		} else {
			b.WriteString(fmt.Sprintf("  %s %s\n", StyleSuccess.Render(IconSuccess), f))
		}
	}
	
	if m.hookCmd != "" {
		b.WriteString(fmt.Sprintf("\n  %s Running hook: %s\n", StyleBrand.Render("❯"), StyleHighlight.Render(m.hookCmd)))
		if m.hookLine != "" {
			b.WriteString(fmt.Sprintf("    %s\n", StyleMuted.Render(m.hookLine)))
		}
	}

	b.WriteString("\n\n")

	percent := 0.0
	if m.total > 0 {
		percent = float64(m.current) / float64(m.total)
	}

	b.WriteString(fmt.Sprintf("  Progress  %s  %d/%d files\n", m.prog.ViewAs(percent), m.current, m.total))

	return b.String()
}

// RunProgress menjalankan animasi progress berdasarkan update channel
func RunProgress(total int, updates <-chan interface{}) error {
	p := tea.NewProgram(initialProgressModel(total, updates))
	_, err := p.Run()
	return err
}

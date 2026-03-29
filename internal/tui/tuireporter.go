package tui

import (
	"fmt"
	"os"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/username/symphony/internal/engine"
)

// TUIReporter implementasi reporter yang memperbarui Bubbletea GUI progress.
type TUIReporter struct {
	totalFiles int
	current    int
	updates    chan interface{}
	p          *tea.Program
	wg         sync.WaitGroup
	mu         sync.Mutex
	closed     bool
}

// NewTUIReporter membuat TUIReporter baru dan langsung memulai Bubbletea program
// di goroutine terpisah. Panggil Wait() untuk menunggu program selesai.
func NewTUIReporter(totalFiles int) *TUIReporter {
	updates := make(chan interface{}, 100)
	model := initialProgressModel(totalFiles, updates)
	p := tea.NewProgram(model)

	r := &TUIReporter{
		totalFiles: totalFiles,
		updates:    updates,
		p:          p,
	}

	// Jalankan TUI di goroutine terpisah; track via WaitGroup
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running progress TUI: %v\n", err)
		}
	}()

	return r
}

func (r *TUIReporter) OnFileCreated(path string) {
	r.mu.Lock()
	r.current++
	current := r.current
	r.mu.Unlock()

	r.send(progressMsg{
		Total:     r.totalFiles,
		Current:   current,
		File:      path,
		Action:    "CREATE",
		Completed: false,
	})
}

func (r *TUIReporter) OnFileSkipped(path string, reason string) {
	// File skip tidak perlu ditampilkan di progress bar (sudah ada di summary)
}

func (r *TUIReporter) OnHookStart(command string) {
	r.send(hookStartMsg{Command: command})
}

func (r *TUIReporter) OnHookOutput(line string) {
	r.send(hookOutputMsg{Line: line})
}

func (r *TUIReporter) OnComplete(stats engine.CompletionStats) {
	// Kirim sinyal selesai lalu tutup channel
	r.send(progressMsg{Completed: true, Total: r.totalFiles})
	r.closeUpdates()
	// Tunggu Bubbletea program benar-benar selesai sebelum return
	r.wg.Wait()
}

func (r *TUIReporter) OnError(err error) {
	fmt.Fprintf(os.Stderr, "\n%s %v\n", StyleDanger.Render(IconError), err)
	// Pastikan Bubbletea berhenti
	r.send(progressMsg{Completed: true, Total: r.totalFiles})
	r.closeUpdates()
	r.wg.Wait()
}

// send mengirim message ke channel dengan safe guard jika channel sudah ditutup.
func (r *TUIReporter) send(msg interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.closed {
		r.updates <- msg
	}
}

// closeUpdates menutup channel updates sekali saja (thread-safe).
func (r *TUIReporter) closeUpdates() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.closed {
		r.closed = true
		close(r.updates)
	}
}

package engine

import (
	"encoding/json"
	"fmt"
	"os"
)

// PrintError menampilkan error dalam format Symphony yang konsisten
func PrintError(title, field, problem, suggestion, format string) {
	if format == "json" {
		errMap := map[string]any{
			"type":       "error",
			"code":       2,
			"field":      field,
			"message":    problem,
			"suggestion": suggestion,
		}
		b, _ := json.Marshal(errMap)
		fmt.Fprintln(os.Stderr, string(b))
		return
	}

	fmt.Fprintf(os.Stderr, "  ✖ %s\n", title)
	fmt.Fprintf(os.Stderr, "  ─────────────────────────────────────\n")
	if field != "" {
		fmt.Fprintf(os.Stderr, "  Field:    %s\n", field)
	}
	fmt.Fprintf(os.Stderr, "  Masalah:  %s\n", problem)
	if suggestion != "" {
		fmt.Fprintf(os.Stderr, "  Saran:    %s\n", suggestion)
	}
	fmt.Fprintf(os.Stderr, "\n")
}

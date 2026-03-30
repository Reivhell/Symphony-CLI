package cliupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

const (
	defaultOwner = "Reivhell"
	defaultRepo  = "Symphony-CLI"
	apiTimeout   = 3 * time.Second
)

// MaybePrintNewerRelease fetches the latest GitHub release tag and prints a one-line
// notice to stderr if it is newer than currentVersion. It is safe to call from a
// goroutine; errors are ignored.
func MaybePrintNewerRelease(currentVersion string) {
	v := strings.TrimSpace(currentVersion)
	if v == "" || v == "dev" {
		return
	}
	current := normalizeSemverTag(v)
	if !semver.IsValid(current) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", defaultOwner, defaultRepo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return
	}
	latest := normalizeSemverTag(strings.TrimSpace(payload.TagName))
	if !semver.IsValid(latest) {
		return
	}
	if semver.Compare(latest, current) > 0 {
		_, _ = fmt.Fprintf(os.Stderr, "symphony: a newer release is available (%s); you are on %s\n", strings.TrimPrefix(latest, "v"), strings.TrimPrefix(current, "v"))
	}
}

func normalizeSemverTag(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "v") {
		s = "v" + s
	}
	return s
}

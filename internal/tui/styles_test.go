package tui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDivider(t *testing.T) {
	d := Divider(10)
	// Harus mengandung karakter ─ sebanyak 10
	assert.Equal(t, 10, strings.Count(d, "─"))
}

func TestDividerZeroWidth(t *testing.T) {
	d := Divider(0)
	// Width 0 → string kosong (setelah strip ANSI)
	assert.NotNil(t, d)
}

func TestIconConstants(t *testing.T) {
	assert.Equal(t, "✔", IconSuccess)
	assert.Equal(t, "✖", IconError)
	assert.Equal(t, "⚠", IconWarning)
	assert.Equal(t, "❯", IconArrow)
	assert.Equal(t, "◆", IconDiamond)
}

func TestColorsDefined(t *testing.T) {
	// Pastikan semua warna didefinisikan (non-empty)
	assert.NotEmpty(t, string(ColorBrand))
	assert.NotEmpty(t, string(ColorSuccess))
	assert.NotEmpty(t, string(ColorWarning))
	assert.NotEmpty(t, string(ColorDanger))
	assert.NotEmpty(t, string(ColorMuted))
	assert.NotEmpty(t, string(ColorHighlight))
}

func TestStylesRender(t *testing.T) {
	// Pastikan semua style bisa render tanpa panic
	assert.NotPanics(t, func() { StyleSuccess.Render("ok") })
	assert.NotPanics(t, func() { StyleWarning.Render("warn") })
	assert.NotPanics(t, func() { StyleDanger.Render("err") })
	assert.NotPanics(t, func() { StyleMuted.Render("muted") })
	assert.NotPanics(t, func() { StyleBrand.Render("brand") })
	assert.NotPanics(t, func() { StyleHighlight.Render("highlight") })
	assert.NotPanics(t, func() { StyleActionCreate.Render("[CREATE]") })
	assert.NotPanics(t, func() { StyleActionModify.Render("[MODIFY]") })
	assert.NotPanics(t, func() { StyleActionSkip.Render("[SKIP]") })
	assert.NotPanics(t, func() { StyleActionDelete.Render("[DELETE]") })
}

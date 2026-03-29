package tui

import "fmt"

// RenderHeader mengembalikan string banner Symphony yang siap ditampilkan.
// version adalah string versi CLI yang diteruskan dari cmd layer.
func RenderHeader(version string) string {
	topLine := fmt.Sprintf("%s Symphony%*sv%s", IconDiamond, 54-1-8-1-len(version), "", version)
	content := topLine + "\n" + "The Adaptive Scaffolding Engine"
	return "\n" + StyleHeader.Render(content) + "\n\n"
}

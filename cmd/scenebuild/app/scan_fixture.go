package app
import (
// 	"fmt"
// 
// 	"github.com/g3n/engine/gui"
	"github.com/tgreiser/cymapper/cmd/scenebuild/ui"
)

type ScanFixture struct {
	app    *App
}
func (s *ScanFixture) Initialize(a *App) {
	a.GuiPanel().Add(ui.NewControlPanel())
}

func (s *ScanFixture) Render(a *App) {
}
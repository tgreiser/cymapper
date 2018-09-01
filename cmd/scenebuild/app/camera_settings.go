package app

import (
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/math32"
)

type CameraSettings struct {
	devId *gui.Edit
}

func (s *CameraSettings) Initialize(a *App) {
	// Adds control panel after the header
	cpanel := gui.NewPanel(800, 120)
	cpanel.SetBorders(0, 0, 1, 0)
	cpanel.SetPaddings(4, 4, 4, 4)
	cpanel.SetColor(math32.NewColorHex(0xffca6e))
	cpanel.SetLayoutParams(&gui.DockLayoutParams{Edge: gui.DockTop})

	// Add GUI stuff
	l := gui.NewLabel("Camera Device ID (0 - ?)")
	l.SetPosition(0, 0)
	l.SetColor(darkTextColor)
	a.Log().Info("Add label")
	cpanel.Add(l)

	s.devId = gui.NewEdit(50, "0")
	s.devId.SetPosition(200, 0)
	a.Log().Info("Add dev id")
	cpanel.Add(s.devId)

	a.GuiPanel().Add(cpanel)
}

func (s *CameraSettings) Render(a *App) {

}

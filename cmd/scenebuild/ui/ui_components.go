package ui

import (
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/math32"
)

func NewControlPanel() *gui.Panel {
	cpanel := gui.NewPanel(800, 120)
	cpanel.SetBorders(0, 0, 1, 0)
	cpanel.SetPaddings(4, 4, 4, 4)
	cpanel.SetColor(math32.NewColorHex(0xffca6e))
	cpanel.SetLayoutParams(&gui.DockLayoutParams{Edge: gui.DockTop})
	return cpanel
}

func SceneDimensionsAndZoom() {

}
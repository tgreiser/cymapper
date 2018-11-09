package app

import (
	"fmt"

	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/window"
)

var darkTextColor = &math32.Color{.4, .4, .4}

// buildGui builds the tester GUI
func (app *App) buildGui() {
	// Create dock layout for the tester root panel
	dl := gui.NewDockLayout()
	app.Gui().SetLayout(dl)

	// Add a transparent panel to contain the canvas
	canvas := gui.NewPanel(0, 0)
	canvas.SetRenderable(false)
	canvas.SetColor4(&gui.StyleDefault().Scroller.BgColor)
	canvas.SetLayoutParams(&gui.DockLayoutParams{Edge: gui.DockCenter})
	canvas.Subscribe(gui.OnMouseDown, func(name string, ev interface{}) {
		mev := ev.(*window.MouseEvent)
		//width, height := app.Window().Size()
		w, h := canvas.Size()
		x := mev.Xpos
		y := mev.Ypos
		app.Log().Debug("%v x %v in window %v x %v\n", x, y, w, h)
	})
	app.Gui().Add(canvas)
	app.SetPanel3D(canvas)

	// Adds header after the gui central panel to ensure that the control folder
	// stays over the gui panel when opened.
	lightTextColor := math32.Color4{0.8, 0.8, 0.8, 1}
	header := gui.NewPanel(600, 30)
	header.SetBorders(0, 0, 1, 0)
	header.SetPaddings(5, 5, 5, 5)
	header.SetColor(math32.NewColorHex(0x956eff))
	header.SetLayoutParams(&gui.DockLayoutParams{Edge: gui.DockTop})

	// Horizontal box layout for the header
	hbox := gui.NewHBoxLayout()
	hbox.SetSpacing(10)
	header.SetLayout(hbox)
	app.Gui().Add(header)

	// Header title
	const fontSize = 20
	title := gui.NewLabel(" ")
	title.SetFontSize(fontSize)
	title.SetLayoutParams(&gui.HBoxLayoutParams{AlignV: gui.AlignTop})
	title.SetText(fmt.Sprintf("%s v%d.%d  ", progName, vmajor, vminor))
	title.SetColor4(&lightTextColor)
	header.Add(title)

	// FPS
	if !*oHideFPS {
		l1 := gui.NewLabel(" ")
		l1.SetFontSize(fontSize)
		l1.SetLayoutParams(&gui.HBoxLayoutParams{AlignV: gui.AlignTop})
		l1.SetText("                 FPS: ")
		l1.SetColor4(&lightTextColor)
		header.Add(l1)
		// FPS value
		app.labelFPS = gui.NewLabel(" ")
		app.labelFPS.SetFontSize(fontSize)
		app.labelFPS.SetLayoutParams(&gui.HBoxLayoutParams{AlignV: gui.AlignTop})
		app.labelFPS.SetColor4(&lightTextColor)
		header.Add(app.labelFPS)
	}

	bTestCam := gui.NewButton("Set Up Camera")
	bTestCam.SetWidth(90)
	bTestCam.SetHeight(30)
	bTestCam.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		app.Log().Info("setupScene")
		app.setupScene()

		switch app.screen.(type) {
		case *SceneUI:
			app.screen = &CameraSettings{}
		default:
			app.screen = &SceneUI{}
		}
		app.screen.Initialize(app)
		app.Log().Info("Toggling scene")
	})
	header.Add(bTestCam)

	bScanFixture := gui.NewButton("Scan Fixture")
	bScanFixture.SetWidth(90)
	bScanFixture.SetHeight(30)
	bScanFixture.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		app.Log().Info("Go to Scan Fixture scene")
		app.setupScene()

		switch app.screen.(type) {
		case *SceneUI:
			app.screen = &ScanFixture{}
		default:
			app.screen = &SceneUI{}
		}
		app.screen.Initialize(app)
		app.Log().Info("Toggling scene")
	})
	header.Add(bScanFixture)

	app.ed = NewErrorDialog(600, 100)
	header.Add(app.ed)
	/*
		err = app.Renderer().AddDefaultShaders()
		if err != nil {
			panic(err)
		}
		app.Renderer().SetScene(app.Scene())
	*/
}

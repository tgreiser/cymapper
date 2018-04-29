package app

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/tgreiser/cymapper/cmd/scenebuild/fixture"
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
	app.Gui().Add(canvas)
	app.SetPanel3D(canvas)

	// Adds header after the gui central panel to ensure that the control folder
	// stays over the gui panel when opened.
	lightTextColor := math32.Color4{0.8, 0.8, 0.8, 1}
	header := gui.NewPanel(600, 40)
	header.SetBorders(0, 0, 1, 0)
	header.SetPaddings(4, 4, 4, 4)
	header.SetColor(math32.NewColorHex(0x956eff))
	header.SetLayoutParams(&gui.DockLayoutParams{Edge: gui.DockTop})

	// Horizontal box layout for the header
	hbox := gui.NewHBoxLayout()
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

	zoom := gui.NewHSlider(100, 30)
	zoom.SetPosition(420, 0)
	zoom.SetText("Zoom")
	zoom.Subscribe(gui.OnChange, func(name string, ev interface{}) {
		app.CameraOrtho().SetZoom(zoom.Value() / 100)
		app.SetCamera(app.CameraOrtho())
	})
	header.Add(zoom)

	// Adds control panel after the header
	cpanel := gui.NewPanel(600, 80)
	cpanel.SetBorders(0, 0, 1, 0)
	cpanel.SetPaddings(4, 4, 4, 4)
	cpanel.SetColor(math32.NewColorHex(0xffca6e))
	cpanel.SetLayoutParams(&gui.DockLayoutParams{Edge: gui.DockTop})

	l2 := gui.NewLabel("Build a scene by adding, moving and resizing fixture maps.")
	l2.SetPosition(10, 10)
	l2.SetPaddings(2, 2, 2, 2)
	l2.SetColor(darkTextColor)
	cpanel.Add(l2)

	fixtures := gui.NewDropDown(200, gui.NewImageLabel(""))
	fixtures.SetHeight(30)
	fixtures.SetPosition(160, 50)

	cpanel.Add(fixtures)
	fixtures.Subscribe(gui.OnChange, func(name string, ev interface{}) {
		app.Log().Debug("Change fixture " + fixtures.Selected().Text())
	})

	bAddFixture := gui.NewButton("Add Fixture")
	bAddFixture.SetPosition(10, 50)
	bAddFixture.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		// browse for file
		app.fs.Show(true)
	})
	cpanel.Add(bAddFixture)

	bReset := gui.NewButton("Reset")
	bReset.SetPosition(100, 50)
	bReset.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		app.Scene().RemoveAll(true)
		app.setupScene()
		fixtures.RemoveAll(false)
		fixtures.Add(gui.NewImageLabel(""))
		fixtures.SelectPos(0)
	})
	cpanel.Add(bReset)

	// Fixture corner controls
	lx := gui.NewLabel("X")
	lx.SetPosition(486, 10)
	lx.SetColor(darkTextColor)
	cpanel.Add(lx)

	ly := gui.NewLabel("Y")
	ly.SetPosition(546, 10)
	ly.SetColor(darkTextColor)
	cpanel.Add(ly)

	tl := gui.NewLabel("Top Left")
	tl.SetPosition(408, 30)
	tl.SetColor(darkTextColor)
	cpanel.Add(tl)

	tlx := gui.NewEdit(50, "")
	tlx.SetPosition(462, 32)
	cpanel.Add(tlx)

	tly := gui.NewEdit(50, "")
	tly.SetPosition(522, 32)
	cpanel.Add(tly)

	br := gui.NewLabel("Bottom Right")
	br.SetPosition(380, 50)
	br.SetColor(darkTextColor)
	cpanel.Add(br)

	brx := gui.NewEdit(50, "")
	brx.SetPosition(462, 52)
	cpanel.Add(brx)

	bry := gui.NewEdit(50, "")
	bry.SetPosition(522, 52)
	cpanel.Add(bry)

	// Creates file selection dialog
	fs, err := NewFileSelect(400, 300)
	if err != nil {
		panic(err)
	}
	app.fs = fs
	app.fs.SetVisible(false)
	app.fs.Subscribe("OnOK", func(evname string, ev interface{}) {
		fpath := app.fs.Selected()
		if fpath == "" {
			app.ed.Show("File not selected")
			return
		}
		app.log.Info("Selected file: %v", fpath)
		// parse relative vectors for fixture
		fixture := fixture.New(fpath)
		mat := material.NewStandard(math32.NewColor("White"))
		mat.SetSide(material.SideDouble)
		mat.SetWireframe(false)
		// add fixture vectors to scene
		for fixture.Available() {
			geom := geometry.NewCircle(3, 16)
			circle := graphic.NewMesh(geom, mat)
			circle.SetPositionVec(fixture.Next())
			app.Scene().Add(circle)
		}
		err = app.Renderer().AddDefaultShaders()
		if err != nil {
			panic(err)
		}
		app.Renderer().SetScene(app.Scene())
		app.fs.Show(false)

		fixtures.Add(gui.NewImageLabel(filepath.Base(fpath)))
		fixtures.SelectPos(fixtures.Len() - 1)

		tlx.SetText(strconv.FormatFloat(float64(fixture.TopLeft().X), 'f', 2, 32))
		tly.SetText(strconv.FormatFloat(float64(fixture.TopLeft().Y), 'f', 2, 32))
		brx.SetText(strconv.FormatFloat(float64(fixture.BottomRight().X), 'f', 2, 32))
		bry.SetText(strconv.FormatFloat(float64(fixture.BottomRight().Y), 'f', 2, 32))
	})
	app.fs.Subscribe("OnCancel", func(evname string, ev interface{}) {
		app.fs.Show(false)
	})
	app.Gui().Add(app.fs)

	app.Gui().Add(cpanel)
}

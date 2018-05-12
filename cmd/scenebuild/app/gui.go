package app

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/window"
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
		app.selected = fixtures.SelectedPos()
		app.Log().Debug("Change fixture %v %v", fixtures.SelectedPos(), fixtures.Selected().Text())
	})

	bAddFixture := gui.NewButton("Add Fixture")
	bAddFixture.SetPosition(10, 50)
	bAddFixture.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		// browse for file
		app.fs.Show(true)
	})
	cpanel.Add(bAddFixture)

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

	wl := gui.NewLabel("Width")
	wl.SetPosition(600, 32)
	wl.SetColor(darkTextColor)
	cpanel.Add(wl)

	we := gui.NewEdit(50, "640")
	we.SetText("640")
	we.SetPosition(650, 32)

	cpanel.Add(we)

	hl := gui.NewLabel("Height")
	hl.SetPosition(600, 52)
	hl.SetColor(darkTextColor)
	cpanel.Add(hl)

	he := gui.NewEdit(50, "480")
	he.SetText("480")
	he.SetPosition(650, 52)
	drawBounds := func(name string, ev interface{}) {
		app.DrawBounds(we.Text(), he.Text())
	}
	we.Subscribe(gui.OnChange, drawBounds)
	he.Subscribe(gui.OnChange, drawBounds)
	cpanel.Add(he)
	app.DrawBounds(we.Text(), he.Text())

	bReset := gui.NewButton("Reset")
	bReset.SetPosition(100, 50)
	bReset.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		app.Scene().RemoveAll(true)
		app.setupScene()
		// TODO - make fixtures reset correctly
		// currently it disappears
		fixtures.SelectPos(-1)
		fixtures.RemoveAll(true)

		app.selected = -1
		app.fixtures = app.fixtures[:0]

		tlx.SetText("")
		tly.SetText("")
		brx.SetText("")
		bry.SetText("")
		drawBounds("", "")
	})
	cpanel.Add(bReset)

	xform := func(name string, ev interface{}) {
		// use app.selected to calculate transformations
		// orig TL - app.fixtures[app.selected].tl
		// orig BR - app.fixtures[app.selected].br
		// new TL - tlx.Text(), tly.Text()
		// new BR - brx.Text(), bry.Text()
		ntlx, err := strconv.ParseFloat(tlx.Text(), 32)
		if err != nil {
			app.Log().Error("Invalid top left coordinates %v\n", tlx.Text())
		}
		ntly, err := strconv.ParseFloat(tly.Text(), 32)
		if err != nil {
			app.Log().Error("Invalid top left coordinates %v\n", tly.Text())
		}
		nbrx, err := strconv.ParseFloat(brx.Text(), 32)
		if err != nil {
			app.Log().Error("Invalid bottom right coordinates %v\n", brx.Text())
		}
		nbry, err := strconv.ParseFloat(bry.Text(), 32)
		if err != nil {
			app.Log().Error("Invalid bottom right coordinates %v\n", bry.Text())
		}

		newTL := math32.NewVector3(float32(ntlx), float32(ntly), 0)
		newBR := math32.NewVector3(float32(nbrx), float32(nbry), 0)
		sc, tr := fixture.NewTransformation(app.fixtures[app.selected].TopLeft(),
			app.fixtures[app.selected].BottomRight(), newTL, newBR)
		app.fixtures[app.selected].Transform(sc, tr)
		app.Log().Debug("SC %v x %v TR %v x %v\n", sc.X, sc.Y, tr.X, tr.Y)

		app.Scene().RemoveAll(false)
		app.setupScene()
		drawBounds("", "")
		app.DrawFixtures()
	}
	tlx.Subscribe(gui.OnChange, xform)
	tly.Subscribe(gui.OnChange, xform)
	brx.Subscribe(gui.OnChange, xform)
	bry.Subscribe(gui.OnChange, xform)

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
		app.fixtures = append(app.fixtures, fixture)
		app.DrawFixtures()
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

	err = app.Renderer().AddDefaultShaders()
	if err != nil {
		panic(err)
	}
	app.Renderer().SetScene(app.Scene())
}

func (app *App) DrawBounds(width, height string) {
	mat := material.NewBasic()

	geom := geometry.NewGeometry()
	vertices := math32.NewArrayF32(0, 16)
	vertices.Append(
		0, 0, 0,
		0, 480, 0,
		0, 480, 0,
		640, 480, 0,
		640, 480, 0,
		640, 0, 0,
		640, 0, 0,
		0, 0, 0,
	)
	colors := math32.NewArrayF32(0, 16)
	colors.Append(
		1, 1, 1,
		1, 1, 1,
		1, 1, 1,
		1, 1, 1,
		1, 1, 1,
		1, 1, 1,
		1, 1, 1,
		1, 1, 1,
	)
	geom.AddVBO(gls.NewVBO().AddAttrib("VertexPosition", 3).SetBuffer(vertices))
	geom.AddVBO(gls.NewVBO().AddAttrib("VertexColor", 3).SetBuffer(colors))

	box := graphic.NewLines(geom, mat)
	app.Scene().Add(box)
}

func (app *App) DrawFixtures() {
	mat := material.NewStandard(math32.NewColor("White"))
	mat.SetSide(material.SideDouble)
	mat.SetWireframe(false)

	rmat := material.NewStandard(math32.NewColor("red"))
	rmat.SetSide(material.SideFront)
	rmat.SetWireframe(true)
	rmat.SetLineWidth(1)

	l := len(app.fixtures)
	for iX := 0; iX < l; iX++ {
		// add fixture vectors to scene
		app.fixtures[iX].Reset()
		for app.fixtures[iX].Available() {
			geom := geometry.NewCircle(3, 16)
			circle := graphic.NewMesh(geom, mat)
			circle.SetPositionVec(app.fixtures[iX].Next())
			app.Scene().Add(circle)
			app.Log().Debug("%v", circle.Position())
		}
		circle := graphic.NewMesh(geometry.NewCircle(6, 16), rmat)
		circle.SetPositionVec(app.fixtures[iX].TopLeft())
		app.Scene().Add(circle)
		circle = graphic.NewMesh(geometry.NewCircle(6, 16), rmat)
		circle.SetPositionVec(app.fixtures[iX].BottomRight())
		app.Scene().Add(circle)
	}
	err := app.Renderer().AddDefaultShaders()
	if err != nil {
		panic(err)
	}
	app.Renderer().SetScene(app.Scene())
}
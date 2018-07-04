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
	header := gui.NewPanel(600, 30)
	header.SetBorders(0, 0, 1, 0)
	header.SetPaddings(5, 5, 5, 5)
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
	l2.SetPosition(0, 0)
	l2.SetPaddings(2, 2, 2, 2)
	l2.SetColor(darkTextColor)
	cpanel.Add(l2)

	fixtures := gui.NewDropDown(200, gui.NewImageLabel(""))
	fixtures.SetHeight(26)
	fixtures.SetPosition(162, 22)

	cpanel.Add(fixtures)
	fixtures.Subscribe(gui.OnChange, func(name string, ev interface{}) {
		//app.selected = fixtures.SelectedPos()
		//app.Log().Debug("Change fixture %v %v", fixtures.SelectedPos(), fixtures.Selected().Text())
		app.Draw()
		app.SetCorners()
	})

	bAddFixture := gui.NewButton("Add Fixture")
	bAddFixture.SetPosition(4, 22)
	bAddFixture.SetWidth(90)
	bAddFixture.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		// browse for file
		app.fs.Show(true)
	})
	cpanel.Add(bAddFixture)

	bSaveScene := gui.NewButton("Save Scene")
	bSaveScene.SetPosition(4, 52)
	bSaveScene.SetWidth(90)
	bSaveScene.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		// save scene dialog
		app.sceneFS.Show(true)
	})
	cpanel.Add(bSaveScene)

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

	app.tlx = gui.NewEdit(50, "")
	app.tlx.SetPosition(462, 32)
	cpanel.Add(app.tlx)

	app.tly = gui.NewEdit(50, "")
	app.tly.SetPosition(522, 32)
	cpanel.Add(app.tly)

	br := gui.NewLabel("Bottom Right")
	br.SetPosition(380, 50)
	br.SetColor(darkTextColor)
	cpanel.Add(br)

	app.brx = gui.NewEdit(50, "")
	app.brx.SetPosition(462, 52)
	cpanel.Add(app.brx)

	app.bry = gui.NewEdit(50, "")
	app.bry.SetPosition(522, 52)
	cpanel.Add(app.bry)

	wl := gui.NewLabel("Width")
	wl.SetPosition(600, 32)
	wl.SetColor(darkTextColor)
	cpanel.Add(wl)

	app.width = gui.NewEdit(50, FormatFloat32(app.sceneWidth))
	app.width.SetText(FormatFloat32(app.sceneWidth))
	app.width.SetPosition(650, 32)

	cpanel.Add(app.width)

	hl := gui.NewLabel("Height")
	hl.SetPosition(600, 52)
	hl.SetColor(darkTextColor)
	cpanel.Add(hl)

	app.height = gui.NewEdit(50, FormatFloat32(app.sceneHeight))
	app.height.SetText(FormatFloat32(app.sceneHeight))
	app.height.SetPosition(650, 52)
	drawBounds := func(name string, ev interface{}) {
		app.sceneWidth = ParseFloat32(app.width.Text(), app.sceneWidth)
		app.sceneHeight = ParseFloat32(app.height.Text(), app.sceneHeight)
		app.Draw()
	}
	app.width.Subscribe(gui.OnChange, drawBounds)
	app.height.Subscribe(gui.OnChange, drawBounds)
	cpanel.Add(app.height)
	app.Draw()

	bReset := gui.NewButton("Reset")
	bReset.SetPosition(98, 22)
	bReset.SetWidth(60)
	bReset.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		app.Scene().RemoveAll(true)
		app.setupScene()
		// TODO - make fixtures reset correctly
		// currently it disappears
		fixtures.SelectPos(-1)
		fixtures.RemoveAll(true)

		app.selected = -1
		app.fixtures = app.fixtures[:0]

		app.tlx.SetText("")
		app.tly.SetText("")
		app.brx.SetText("")
		app.bry.SetText("")
		drawBounds("", "")
	})
	cpanel.Add(bReset)

	xform := func(name string, ev interface{}) {
		// use app.selected to calculate transformations
		// orig TL - app.fixtures[app.selected].tl
		// orig BR - app.fixtures[app.selected].br
		// new TL - tlx.Text(), tly.Text()
		// new BR - brx.Text(), bry.Text()
		ntlx, err := strconv.ParseFloat(app.tlx.Text(), 32)
		if err != nil {
			app.Log().Error("Invalid top left coordinates %v\n", app.tlx.Text())
		}
		ntly, err := strconv.ParseFloat(app.tly.Text(), 32)
		if err != nil {
			app.Log().Error("Invalid top left coordinates %v\n", app.tly.Text())
		}
		nbrx, err := strconv.ParseFloat(app.brx.Text(), 32)
		if err != nil {
			app.Log().Error("Invalid bottom right coordinates %v\n", app.brx.Text())
		}
		nbry, err := strconv.ParseFloat(app.bry.Text(), 32)
		if err != nil {
			app.Log().Error("Invalid bottom right coordinates %v\n", app.bry.Text())
		}

		newTL := math32.NewVector3(float32(ntlx), float32(ntly), 0)
		newBR := math32.NewVector3(float32(nbrx), float32(nbry), 0)
		sc, tr := fixture.NewTransformation(app.fixtures[app.selected].TopLeft(),
			app.fixtures[app.selected].BottomRight(), newTL, newBR)
		app.fixtures[app.selected].Transform(sc, tr)
		app.Log().Debug("selected %v x %v\n", app.fixtures[app.selected].TopLeft(), app.fixtures[app.selected].BottomRight())
		app.Log().Debug("SC %v x %v TR %v x %v\n", sc.X, sc.Y, tr.X, tr.Y)

		app.Draw()
	}
	app.tlx.Subscribe(gui.OnChange, xform)
	app.tly.Subscribe(gui.OnChange, xform)
	app.brx.Subscribe(gui.OnChange, xform)
	app.bry.Subscribe(gui.OnChange, xform)

	// Save Scene - File Select
	ss, err := NewFileSelect(400, 300)
	if err != nil {
		panic(err)
	}
	app.sceneFS = ss
	app.sceneFS.SetVisible(false)
	app.sceneFS.SetTitle("Save Scene")
	app.sceneFS.Subscribe("OnOK", func(evname string, ev interface{}) {
		fpath := app.sceneFS.Selected()
		if fpath == "" {
			app.ed.Show("Please enter a path for your scene")
			return
		}
		app.log.Info("Selected file: %v", fpath)
		app.sceneFS.Show(false)
		// write all the fixtures merged into a single TSV
		scene := fixture.NewScene(app.fixtures)
		scene.SaveAs(fpath)
	})
	app.sceneFS.Subscribe("OnCancel", func(evname string, ev interface{}) {
		app.sceneFS.Show(false)
	})
	app.Gui().Add(app.sceneFS)

	// Add Fixture - File Select
	fs, err := NewFileSelect(400, 300)
	if err != nil {
		panic(err)
	}
	app.fs = fs
	app.fs.SetVisible(false)
	app.fs.SetTitle("Add Fixture")
	app.fs.Subscribe("OnOK", func(evname string, ev interface{}) {
		fpath := app.fs.Selected()
		if fpath == "" {
			app.ed.Show("File not selected")
			return
		}
		app.log.Info("Selected file: %v", fpath)
		// parse relative vectors for fixture
		fixture.NewFixture(fpath, app)
        app.DrawFixtures()
		app.fs.Show(false)

		fixtures.Add(gui.NewImageLabel(filepath.Base(fpath)))
		fixtures.SelectPos(fixtures.Len() - 1)

		app.SetCorners()
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

func (app *App) SetCorners() {
	fixture := app.fixtures[app.selected]
	app.tlx.SetText(FormatFloat32(fixture.TransformedTopLeft().X))
	app.tly.SetText(FormatFloat32(fixture.TransformedTopLeft().Y))
	app.brx.SetText(FormatFloat32(fixture.TransformedBottomRight().X))
	app.bry.SetText(FormatFloat32(fixture.TransformedBottomRight().Y))
}

func (app *App) Draw() {
	app.Scene().RemoveAll(false)
	app.setupScene()
	app.DrawBounds()
	app.DrawFixtures()
}

func (app *App) DrawBounds() {
	mat := material.NewBasic()

	gmat := material.NewStandard(math32.NewColor("green"))
	gmat.SetSide(material.SideFront)
	gmat.SetWireframe(true)
	gmat.SetLineWidth(1)

	geom := geometry.NewCircle(1, 16)
	circle := graphic.NewMesh(geom, gmat)
	circle.SetPosition(app.sceneWidth/2, app.sceneHeight/2, 0)
	app.Scene().Add(circle)

	geom2 := geometry.NewGeometry()
	vertices := math32.NewArrayF32(0, 16)
	vertices.Append(
		0, 0, 0,
		0, app.sceneHeight, 0,
		0, app.sceneHeight, 0,
		app.sceneWidth, app.sceneHeight, 0,
		app.sceneWidth, app.sceneHeight, 0,
		app.sceneWidth, 0, 0,
		app.sceneWidth, 0, 0,
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
	geom2.AddVBO(gls.NewVBO().AddAttrib("VertexPosition", 3).SetBuffer(vertices))
	geom2.AddVBO(gls.NewVBO().AddAttrib("VertexColor", 3).SetBuffer(colors))

	box := graphic.NewLines(geom2, mat)
	app.Scene().Add(box)
}

// Will be unneccessary when NewFixture is complete
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
			//app.Log().Debug("%v", circle.Position())
		}
		if iX == app.selected {
			circle := graphic.NewMesh(geometry.NewCircle(6, 16), rmat)
			circle.SetPositionVec(app.fixtures[iX].TransformedTopLeft())
			app.Scene().Add(circle)
			circle = graphic.NewMesh(geometry.NewCircle(6, 16), rmat)
			circle.SetPositionVec(app.fixtures[iX].TransformedBottomRight())
			app.Scene().Add(circle)
		}
	}
	err := app.Renderer().AddDefaultShaders()
	if err != nil {
		panic(err)
	}
	app.Renderer().SetScene(app.Scene())
}

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
	"github.com/gerow/go-color"
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

	app.zoom = gui.NewHSlider(100, 30)
	app.zoom.SetPosition(420, 0)
	app.zoom.SetText("Zoom")
	app.zoom.SetValue(0.3)
	app.zoom.Subscribe(gui.OnChange, func(name string, ev interface{}) {
		app.CameraOrtho().SetZoom(app.zoom.Value() / 100)
		app.SetCamera(app.CameraOrtho())
	})
	header.Add(app.zoom)

	// Adds control panel after the header
	cpanel := gui.NewPanel(600, 120)
	cpanel.SetBorders(0, 0, 1, 0)
	cpanel.SetPaddings(4, 4, 4, 4)
	cpanel.SetColor(math32.NewColorHex(0xffca6e))
	cpanel.SetLayoutParams(&gui.DockLayoutParams{Edge: gui.DockTop})

	l2 := gui.NewLabel("Build a scene by adding, moving and resizing fixture maps.")
	l2.SetPosition(0, 0)
	l2.SetPaddings(2, 2, 2, 2)
	l2.SetColor(darkTextColor)
	cpanel.Add(l2)

	fixtures := app.newFixturesDropDown(cpanel)

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
		fixtures.SelectPos(-1)

		// Removes and then creates new fixture panel because it's a pain to modify
		cpanel.Remove(fixtures)
		fixtures = app.newFixturesDropDown(cpanel)

		app.selected = -1
		app.fixtures = nil

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

		app.transformFixtureTo(app.CurrentFixture(), ntlx, ntly, nbrx, nbry)
	}
	app.tlx.Subscribe(gui.OnChange, xform)
	app.tly.Subscribe(gui.OnChange, xform)
	app.brx.Subscribe(gui.OnChange, xform)
	app.bry.Subscribe(gui.OnChange, xform)

	bFlipX := gui.NewButton("Flip X")
	bFlipX.SetPosition(454, 80)
	bFlipX.SetWidth(60)
	bFlipX.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		app.flip("X")
	})
	cpanel.Add(bFlipX)

	bFlipY := gui.NewButton("Flip Y")
	bFlipY.SetPosition(522, 80)
	bFlipY.SetWidth(60)
	bFlipY.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		app.flip("Y")
	})
	cpanel.Add(bFlipY)

	// Save Scene - File Select
	ss, err := NewFileSelect(400, 300, "../../fixtures")
	if err != nil {
		panic(err)
	}
	app.sceneFS = ss
	app.sceneFS.SetVisible(false)
	app.sceneFS.SetTitle("Save Scene")
	app.sceneFS.Subscribe("OnOK", func(evname string, ev interface{}) {
		fpath, err := app.sceneFS.Selected()
		if err != nil {
			if err.Error() == "file not selected" {
				app.ed.Show("File not selected")
			}
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
	fs, err := NewFileSelect(400, 300, "../../fixtures")
	if err != nil {
		panic(err)
	}
	app.fs = fs
	app.fs.SetVisible(false)
	app.fs.SetTitle("Add Fixture")
	app.fs.Subscribe("OnOK", func(evname string, ev interface{}) {
		fpath, err := app.fs.Selected()
		if err != nil {
			if err.Error() == "file not selected" {
				app.ed.Show("File not selected")
			}
			return
		}
		app.log.Info("Selected file: %v", fpath)
		// parse relative vectors for fixture
		app.newFixture(fpath)
		app.DrawFixtures()
		app.fs.Show(false)
		newFixture := gui.NewImageLabel(filepath.Base(fpath))
		fixtures.Add(newFixture)
		fixtures.SelectPos(fixtures.Len() - 1)
		app.selected = fixtures.Len() - 1

		app.SetCorners()
		app.Draw()
	})
	app.fs.Subscribe("OnCancel", func(evname string, ev interface{}) {
		app.fs.Show(false)
	})
	app.Gui().Add(app.fs)

	app.ed = NewErrorDialog(600, 100)
	app.Gui().Add(app.ed)

	app.Gui().Add(cpanel)

	err = app.Renderer().AddDefaultShaders()
	if err != nil {
		panic(err)
	}
	app.Renderer().SetScene(app.Scene())
}

func (app *App) newFixturesDropDown(cpanel *gui.Panel) *gui.DropDown {
	fixtures := gui.NewDropDown(200, gui.NewImageLabel(""))
	fixtures.SetHeight(26)
	fixtures.SetPosition(162, 22)
	fixtures.SelectPos(-1)

	cpanel.Add(fixtures)
	fixtures.Subscribe(gui.OnChange, func(name string, ev interface{}) {
		app.selected = fixtures.SelectedPos()
		//app.Log().Debug("Change fixture %v %v", fixtures.SelectedPos(), fixtures.Selected().Text())
		app.Draw()
		app.SetCorners()
	})
	return fixtures
}

func (app *App) transformFixtureTo(fixt *fixture.Fixture, ntlx, ntly, nbrx, nbry float64) {
	newTL := math32.NewVector3(float32(ntlx), float32(ntly), 0)
	newBR := math32.NewVector3(float32(nbrx), float32(nbry), 0)
	sc, tr := fixture.NewTransformation(fixt.TopLeft(),
		fixt.BottomRight(), newTL, newBR)
	fixt.Transform(sc, tr)
	app.Log().Debug("selected %v x %v\n", fixt.TopLeft(), fixt.BottomRight())
	app.Log().Debug("SC %v x %v TR %v x %v\n", sc.X, sc.Y, tr.X, tr.Y)

	app.Draw()
}

func (app *App) flip(direction string) {
	topLeftX, err := strconv.ParseFloat(app.tlx.Text(), 32)
	if err != nil {
		app.Log().Error("Invalid top left coordinates %v\n", app.tlx.Text())
	}
	topLeftY, err := strconv.ParseFloat(app.tly.Text(), 32)
	if err != nil {
		app.Log().Error("Invalid top left coordinates %v\n", app.tlx.Text())
	}
	bottomRightX, err := strconv.ParseFloat(app.brx.Text(), 32)
	if err != nil {
		app.Log().Error("Invalid bottom right coordinates %v\n", app.brx.Text())
	}
	bottomRightY, err := strconv.ParseFloat(app.bry.Text(), 32)
	if err != nil {
		app.Log().Error("Invalid bottom right coordinates %v\n", app.brx.Text())
	}

	if app.selected < 0 {
		return
	}
	currentFixture := app.CurrentFixture()
	if direction == "X" {
		// Swap Y values of current fixture.
		app.transformFixtureTo(currentFixture, topLeftX, bottomRightY, bottomRightX, topLeftY)
	} else if direction == "Y" {
		// Swap X values of current fixture.
		app.transformFixtureTo(currentFixture, bottomRightX, topLeftY, topLeftX, bottomRightY)
	} else {
		return
	}

	currentFixture.UpdatePoints()
	// currentFixture.tl, currentFixture.br = currentFixture.FindCorners(currentFixture.pts)

	// app.tly.SetText(bottomRightY)
	// app.bry.SetText(topLeftY)
	// app.tly.Dispatch(gui.OnChange, nil)
	// app.bry.Dispatch(gui.OnChange, nil)
	app.Draw()
}

func (app *App) newFixture(filePath string) {
	newFixture := fixture.NewFixture(filePath)
	app.fixtures = append(app.fixtures, newFixture)
}

func (app *App) SetCorners() {
	if app.selected >= 0 {
		fixture := app.fixtures[app.selected]
		app.tlx.SetText(FormatFloat32(fixture.TransformedTopLeft().X))
		app.tly.SetText(FormatFloat32(fixture.TransformedTopLeft().Y))
		app.brx.SetText(FormatFloat32(fixture.TransformedBottomRight().X))
		app.bry.SetText(FormatFloat32(fixture.TransformedBottomRight().Y))
	}
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
	geom2.AddVBO(gls.NewVBO(vertices).AddAttrib(gls.VertexPosition))
	geom2.AddVBO(gls.NewVBO(colors).AddAttrib(gls.VertexColor))

	box := graphic.NewLines(geom2, mat)
	app.Scene().Add(box)
}

func (app *App) NewRainbowMaterial(hue float64) *material.Standard {
	hslColor := color.HSL{hue, 1.0, 0.5}
	goColorRGB := hslColor.ToRGB()
	g3nRGB := &math32.Color{float32(goColorRGB.R),
		float32(goColorRGB.G),
		float32(goColorRGB.B),
	}

	mat := material.NewStandard(g3nRGB)
	mat.SetSide(material.SideDouble)
	mat.SetWireframe(false)
	return mat
}

func (app *App) DrawFixtures() {

	rmat := material.NewStandard(math32.NewColor("red"))
	rmat.SetSide(material.SideFront)
	rmat.SetWireframe(true)
	rmat.SetLineWidth(1)

	for iX, fixture := range app.fixtures {
		// add fixture vectors to scene
		fixture.Reset()

		for j := 0; fixture.Available(); j++ {
			geom := geometry.NewCircle(3, 16)
			mat := app.NewRainbowMaterial(float64(j) / float64(fixture.Length()) * 0.67)
			circle := graphic.NewMesh(geom, mat)
			circle.SetPositionVec(fixture.Next())
			app.Scene().Add(circle)
			//app.Log().Debug("%v", circle.Position())
		}
		if iX == app.selected {
			circle := graphic.NewMesh(geometry.NewCircle(6, 16), rmat)
			circle.SetPositionVec(fixture.TransformedTopLeft())
			app.Scene().Add(circle)
			circle = graphic.NewMesh(geometry.NewCircle(6, 16), rmat)
			circle.SetPositionVec(fixture.TransformedBottomRight())
			app.Scene().Add(circle)
		}
	}

	err := app.Renderer().AddDefaultShaders()
	if err != nil {
		panic(err)
	}
	app.Renderer().SetScene(app.Scene())
}

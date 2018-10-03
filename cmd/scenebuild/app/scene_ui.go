package app

import (
	"path/filepath"
	"strconv"

	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/util/logger"
	color "github.com/gerow/go-color"
	"github.com/tgreiser/cymapper/cmd/scenebuild/fixture"
)

type SceneUI struct {
	devId       *gui.Edit
	fs          *FileSelect // File select dialog
	sceneFS     *FileSelect
	fixtures    []*fixture.Fixture
	selected    int // selected fixture
	sceneWidth  float32
	sceneHeight float32
	width       *gui.Edit
	height      *gui.Edit
	tlx         *gui.Edit // Top left x; x coordinate of top left corner of current fixture
	tly         *gui.Edit // Top left y
	brx         *gui.Edit // Bottom right x
	bry         *gui.Edit // Bottom right y
	log         *logger.Logger
	app         *App
}

func (s *SceneUI) Initialize(app *App) {
	s.log = app.Log()
	s.app = app

	s.sceneWidth = 1280
	s.sceneHeight = 720
	s.selected = -1

	// Adds control panel after the header
	cpanel := gui.NewPanel(800, 120)
	cpanel.SetBorders(0, 0, 1, 0)
	cpanel.SetPaddings(4, 4, 4, 4)
	cpanel.SetColor(math32.NewColorHex(0xffca6e))
	cpanel.SetLayoutParams(&gui.DockLayoutParams{Edge: gui.DockTop})

	l2 := gui.NewLabel("Build a scene by adding, moving and resizing fixture maps.")
	l2.SetPosition(0, 0)
	l2.SetPaddings(2, 2, 2, 2)
	l2.SetColor(darkTextColor)
	cpanel.Add(l2)

	fixtures := s.newFixturesDropDown(cpanel)

	bAddFixture := gui.NewButton("Add Fixture")
	bAddFixture.SetPosition(4, 22)
	bAddFixture.SetWidth(90)
	bAddFixture.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		// browse for file
		s.fs.Show(true)
	})
	cpanel.Add(bAddFixture)

	bSaveScene := gui.NewButton("Save Scene")
	bSaveScene.SetPosition(4, 52)
	bSaveScene.SetWidth(90)
	bSaveScene.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		// save scene dialog
		s.sceneFS.Show(true)
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

	s.tlx = gui.NewEdit(50, "")
	s.tlx.SetPosition(462, 32)
	cpanel.Add(s.tlx)

	s.tly = gui.NewEdit(50, "")
	s.tly.SetPosition(522, 32)
	cpanel.Add(s.tly)

	br := gui.NewLabel("Bottom Right")
	br.SetPosition(380, 50)
	br.SetColor(darkTextColor)
	cpanel.Add(br)

	s.brx = gui.NewEdit(50, "")
	s.brx.SetPosition(462, 52)
	cpanel.Add(s.brx)

	s.bry = gui.NewEdit(50, "")
	s.bry.SetPosition(522, 52)
	cpanel.Add(s.bry)

	wl := gui.NewLabel("Width")
	wl.SetPosition(600, 32)
	wl.SetColor(darkTextColor)
	cpanel.Add(wl)

	s.width = gui.NewEdit(50, FormatFloat32(s.sceneWidth))
	s.width.SetText(FormatFloat32(s.sceneWidth))
	s.width.SetPosition(650, 32)

	cpanel.Add(s.width)

	hl := gui.NewLabel("Height")
	hl.SetPosition(600, 52)
	hl.SetColor(darkTextColor)
	cpanel.Add(hl)

	s.height = gui.NewEdit(50, FormatFloat32(s.sceneHeight))
	s.height.SetText(FormatFloat32(s.sceneHeight))
	s.height.SetPosition(650, 52)
	drawBounds := func(name string, ev interface{}) {
		s.Draw()
	}
	s.width.Subscribe(gui.OnChange, drawBounds)
	s.height.Subscribe(gui.OnChange, drawBounds)
	cpanel.Add(s.height)
	s.Draw()

	bReset := gui.NewButton("Reset")
	bReset.SetPosition(98, 22)
	bReset.SetWidth(60)
	bReset.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		app.Scene().RemoveAll(true)
		//app.setupScene()
		fixtures.SelectPos(-1)

		// Removes and then creates new fixture panel because it's a pain to modify
		cpanel.Remove(fixtures)
		fixtures = s.newFixturesDropDown(cpanel)

		s.selected = -1
		s.fixtures = nil

		s.tlx.SetText("")
		s.tly.SetText("")
		s.brx.SetText("")
		s.bry.SetText("")
		s.Draw()
	})
	cpanel.Add(bReset)

	xform := func(name string, ev interface{}) {
		// use app.selected to calculate transformations
		// orig TL - app.fixtures[app.selected].tl
		// orig BR - app.fixtures[app.selected].br
		// new TL - tlx.Text(), tly.Text()
		// new BR - brx.Text(), bry.Text()
		ntlx, err := strconv.ParseFloat(s.tlx.Text(), 32)
		if err != nil {
			app.Log().Error("Invalid top left coordinates %v\n", s.tlx.Text())
		}
		ntly, err := strconv.ParseFloat(s.tly.Text(), 32)
		if err != nil {
			app.Log().Error("Invalid top left coordinates %v\n", s.tly.Text())
		}
		nbrx, err := strconv.ParseFloat(s.brx.Text(), 32)
		if err != nil {
			app.Log().Error("Invalid bottom right coordinates %v\n", s.brx.Text())
		}
		nbry, err := strconv.ParseFloat(s.bry.Text(), 32)
		if err != nil {
			app.Log().Error("Invalid bottom right coordinates %v\n", s.bry.Text())
		}

		s.transformFixtureTo(s.CurrentFixture(), ntlx, ntly, nbrx, nbry)
	}
	s.tlx.Subscribe(gui.OnChange, xform)
	s.tly.Subscribe(gui.OnChange, xform)
	s.brx.Subscribe(gui.OnChange, xform)
	s.bry.Subscribe(gui.OnChange, xform)

	bFlipX := gui.NewButton("Flip X")
	bFlipX.SetPosition(454, 80)
	bFlipX.SetWidth(60)
	bFlipX.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		s.flip("X")
	})
	cpanel.Add(bFlipX)

	bFlipY := gui.NewButton("Flip Y")
	bFlipY.SetPosition(522, 80)
	bFlipY.SetWidth(60)
	bFlipY.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		s.flip("Y")
	})
	cpanel.Add(bFlipY)

	// Save Scene - File Select
	ss, err := NewFileSelect(400, 300, "../../fixtures")
	if err != nil {
		panic(err)
	}
	s.sceneFS = ss
	s.sceneFS.SetVisible(false)
	s.sceneFS.SetTitle("Save Scene")
	s.sceneFS.Subscribe("OnOK", func(evname string, ev interface{}) {
		fpath, err := s.sceneFS.Selected()
		if err != nil {
			if err.Error() == "file not selected" {
				app.ed.Show("File not selected")
			}
			return
		}
		app.log.Info("Selected file: %v", fpath)
		s.sceneFS.Show(false)
		// write all the fixtures merged into a single TSV
		scene := fixture.NewScene(s.fixtures)
		scene.SaveAs(fpath)
	})
	s.sceneFS.Subscribe("OnCancel", func(evname string, ev interface{}) {
		s.sceneFS.Show(false)
	})
	cpanel.Add(s.sceneFS)

	// Add Fixture - File Select
	fs, err := NewFileSelect(400, 300, "../../fixtures")
	if err != nil {
		panic(err)
	}
	s.fs = fs
	s.fs.SetVisible(false)
	s.fs.SetTitle("Add Fixture")
	s.fs.Subscribe("OnOK", func(evname string, ev interface{}) {
		fpath, err := s.fs.Selected()
		if err != nil {
			if err.Error() == "file not selected" {
				app.ed.Show("File not selected")
			}
			return
		}
		app.log.Info("Selected file: %v", fpath)
		// parse relative vectors for fixture
		s.newFixture(fpath)
		s.DrawFixtures()
		s.fs.Show(false)
		newFixture := gui.NewImageLabel(filepath.Base(fpath))
		fixtures.Add(newFixture)
		fixtures.SelectPos(fixtures.Len() - 1)
		s.selected = fixtures.Len() - 1

		s.SetCorners()
		s.Draw()
	})
	s.fs.Subscribe("OnCancel", func(evname string, ev interface{}) {
		s.fs.Show(false)
	})
	cpanel.Add(s.fs)

	app.GuiPanel().Add(cpanel)
}

func (s *SceneUI) Render(a *App) {
}

func (s *SceneUI) newFixturesDropDown(cpanel *gui.Panel) *gui.DropDown {
	fixtures := gui.NewDropDown(200, gui.NewImageLabel(""))
	fixtures.SetHeight(26)
	fixtures.SetPosition(162, 22)
	fixtures.SelectPos(-1)

	cpanel.Add(fixtures)
	fixtures.Subscribe(gui.OnChange, func(name string, ev interface{}) {
		s.selected = fixtures.SelectedPos()
		//app.Log().Debug("Change fixture %v %v", fixtures.SelectedPos(), fixtures.Selected().Text())
		s.Draw()
		s.SetCorners()
	})
	return fixtures
}

func (s *SceneUI) transformFixtureTo(fixt *fixture.Fixture, ntlx, ntly, nbrx, nbry float64) {
	newTL := math32.NewVector3(float32(ntlx), float32(ntly), 0)
	newBR := math32.NewVector3(float32(nbrx), float32(nbry), 0)
	sc, tr := fixture.NewTransformation(fixt.TopLeft(),
		fixt.BottomRight(), newTL, newBR)
	fixt.Transform(sc, tr)
	s.Log().Debug("selected %v x %v\n", fixt.TopLeft(), fixt.BottomRight())
	s.Log().Debug("SC %v x %v TR %v x %v\n", sc.X, sc.Y, tr.X, tr.Y)

	s.Draw()
}

func (s *SceneUI) flip(direction string) {
	topLeftX, err := strconv.ParseFloat(s.tlx.Text(), 32)
	if err != nil {
		s.Log().Error("Invalid top left coordinates %v\n", s.tlx.Text())
	}
	topLeftY, err := strconv.ParseFloat(s.tly.Text(), 32)
	if err != nil {
		s.Log().Error("Invalid top left coordinates %v\n", s.tlx.Text())
	}
	bottomRightX, err := strconv.ParseFloat(s.brx.Text(), 32)
	if err != nil {
		s.Log().Error("Invalid bottom right coordinates %v\n", s.brx.Text())
	}
	bottomRightY, err := strconv.ParseFloat(s.bry.Text(), 32)
	if err != nil {
		s.Log().Error("Invalid bottom right coordinates %v\n", s.brx.Text())
	}

	if s.selected < 0 {
		return
	}
	currentFixture := s.CurrentFixture()
	if direction == "X" {
		// Swap Y values of current fixture.
		s.transformFixtureTo(currentFixture, topLeftX, bottomRightY, bottomRightX, topLeftY)
	} else if direction == "Y" {
		// Swap X values of current fixture.
		s.transformFixtureTo(currentFixture, bottomRightX, topLeftY, topLeftX, bottomRightY)
	} else {
		return
	}

	currentFixture.UpdatePoints()
	// currentFixture.tl, currentFixture.br = currentFixture.FindCorners(currentFixture.pts)

	// app.tly.SetText(bottomRightY)
	// app.bry.SetText(topLeftY)
	// app.tly.Dispatch(gui.OnChange, nil)
	// app.bry.Dispatch(gui.OnChange, nil)
	s.Draw()
}

func (s *SceneUI) newFixture(filePath string) {
	newFixture := fixture.NewFixture(filePath)
	s.fixtures = append(s.fixtures, newFixture)
}

func (s *SceneUI) SetCorners() {
	if s.selected >= 0 {
		fixture := s.fixtures[s.selected]
		s.tlx.SetText(FormatFloat32(fixture.TransformedTopLeft().X))
		s.tly.SetText(FormatFloat32(fixture.TransformedTopLeft().Y))
		s.brx.SetText(FormatFloat32(fixture.TransformedBottomRight().X))
		s.bry.SetText(FormatFloat32(fixture.TransformedBottomRight().Y))
	}
}

func (s *SceneUI) Draw() {
	s.app.Scene().RemoveAll(true)
	s.app.Scene().Add(s.app.ambLight)
	s.app.Scene().Add(s.app.CameraOrtho().GetCamera())
	//s.app.setupScene()
	s.CenterCamera()
	s.DrawBounds()
	s.DrawFixtures()
}

func (s *SceneUI) CenterCamera() {
	s.sceneWidth = ParseFloat32(s.width.Text(), s.sceneWidth)
	s.sceneHeight = ParseFloat32(s.height.Text(), s.sceneHeight)
	vx := s.sceneWidth / 2
	vy := s.sceneHeight / 2
	s.app.Log().Debug("CenterCamera %v x %v", vx, vy)
	s.app.CameraOrtho().SetPosition(vx, vy, 99)
	s.app.CameraOrtho().LookAt(&math32.Vector3{vx, vy, 0})
	if s.app.zoom != nil {
		s.app.CameraOrtho().SetZoom(s.app.zoom.Value() / 100)
	}
}

func (s *SceneUI) DrawBounds() {
	mat := material.NewBasic()

	gmat := material.NewStandard(math32.NewColor("green"))
	gmat.SetSide(material.SideFront)
	gmat.SetWireframe(true)
	gmat.SetLineWidth(1)

	geom := geometry.NewCircle(1, 16)
	circle := graphic.NewMesh(geom, gmat)
	circle.SetPosition(s.sceneWidth/2, s.sceneHeight/2, 0)
	s.app.Scene().Add(circle)

	geom2 := geometry.NewGeometry()
	vertices := math32.NewArrayF32(0, 16)
	vertices.Append(
		0, 0, 0,
		0, s.sceneHeight, 0,
		0, s.sceneHeight, 0,
		s.sceneWidth, s.sceneHeight, 0,
		s.sceneWidth, s.sceneHeight, 0,
		s.sceneWidth, 0, 0,
		s.sceneWidth, 0, 0,
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
	s.app.Scene().Add(box)
}

func (s *SceneUI) NewRainbowMaterial(hue float64) *material.Standard {
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

func (s *SceneUI) DrawFixtures() {
	rmat := material.NewStandard(math32.NewColor("red"))
	rmat.SetSide(material.SideFront)
	rmat.SetWireframe(true)
	rmat.SetLineWidth(1)

	for iX, fixture := range s.fixtures {
		// add fixture vectors to scene
		fixture.Reset()
		s.app.Log().Debug("fixture %v", iX)

		for j := 0; fixture.Available(); j++ {
			geom := geometry.NewCircle(3, 16)
			mat := s.NewRainbowMaterial(float64(j) / float64(fixture.Length()) * 0.67)
			circle := graphic.NewMesh(geom, mat)
			circle.SetPositionVec(fixture.Next())
			s.app.Scene().Add(circle)
			//s.app.Log().Debug("%v", circle.Position())
		}
		if iX == s.selected {
			circle := graphic.NewMesh(geometry.NewCircle(6, 16), rmat)
			circle.SetPositionVec(fixture.TransformedTopLeft())
			s.app.Scene().Add(circle)
			circle = graphic.NewMesh(geometry.NewCircle(6, 16), rmat)
			circle.SetPositionVec(fixture.TransformedBottomRight())
			s.app.Scene().Add(circle)
		}
	}

	err := s.app.Renderer().AddDefaultShaders()
	if err != nil {
		panic(err)
	}
	s.app.Renderer().SetScene(s.app.Scene())
}

func (s *SceneUI) CurrentFixture() *fixture.Fixture {
	return s.fixtures[s.selected]
}

func (s *SceneUI) Log() *logger.Logger {
	return s.log
}

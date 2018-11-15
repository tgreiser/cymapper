package app

import (
	"fmt"
	"image"
	"os"
	"strconv"

	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/texture"
	"gocv.io/x/gocv"
	"github.com/tgreiser/cymapper/cmd/scenebuild/ui"
)

type CameraSettings struct {
	devId             *gui.Edit
	deviceId          int
	c                 chan os.Signal
	mat               gocv.Mat
	window            *gocv.Window
	webcam            *gocv.VideoCapture
	webcamAspectRatio float32
	texture           *texture.Texture2D
	app               *App
}

func (s *CameraSettings) Initialize(a *App) {
	s.app = a
	a.Scene().Add(a.ambLight)
	a.CameraOrtho().SetZoom(3)
	a.CameraOrtho().SetPositionY(0.14)
	a.Scene().Add(a.CameraOrtho().GetCamera())
	s.deviceId = 0

	// open webcam
	var err error
	s.webcam, err = gocv.VideoCaptureDevice(s.deviceId)
	if err != nil {
		fmt.Printf("error opening video capture device: %v\n", s.deviceId)
		return
	}

	// prepare image matricies
	s.mat = gocv.NewMat()

	a.AddFinalizer(func() {
		// finalizer will close image and webcam
		s.webcam.Close()
		s.mat.Close()
	})

	cpanel := ui.NewControlPanel()

	// Add GUI stuff
	l := gui.NewLabel("Camera Device ID (0 - ?)")
	l.SetPosition(0, 0)
	l.SetColor(darkTextColor)
	a.Log().Info("Add label")
	cpanel.Add(l)

	s.devId = gui.NewEdit(50, "0")
	s.devId.SetPosition(200, 0)
	s.devId.Subscribe(gui.OnChange, func(name string, ev interface{}) {
		var err error

		oldId, err := strconv.Atoi(s.devId.Text())
		fmt.Printf("old id: %v", oldId)
		s.deviceId, err = strconv.Atoi(s.devId.Text())
		if err != nil {
			fmt.Printf("invalid webcam device Id from gui")
			return
		}
		s.webcam, err = gocv.VideoCaptureDevice(s.deviceId)
		// Error handling not working correctly, error is still nil
		// when an invalid deviceId is chosen, see:
		// https://github.com/hybridgroup/gocv/issues/274
		//if err != nil {
		//	fmt.Printf("error opening video capture device: %v\n", s.deviceId)
		//	s.deviceId = oldId
		//	s.devId.SetText(string(oldId)) 
		//	s.webcam, err = gocv.VideoCaptureDevice(s.deviceId)
		//}
	})
	a.Log().Info("Add dev id")
	cpanel.Add(s.devId)

	a.GuiPanel().Add(cpanel)

	s.makeWebcamView(a)

	s.addGuidelines(a)

	// gocv logic
	// channel to receive os signal
	//s.c = make(chan os.Signal, 1)
	//signal.Notify(s.c, os.Interrupt)
}

func (s *CameraSettings) Render(a *App) {
	imageRGBA := s.getRGBAImageFromWebcam()
	s.texture.SetFromRGBA(imageRGBA)
}


func (s *CameraSettings) makeWebcamView(a *App) {
	imageRGBA := s.getRGBAImageFromWebcam()
	bounds := imageRGBA.Bounds()
	width := float32(bounds.Dx())
	height := float32(bounds.Dy())
	s.webcamAspectRatio = width / height

	s.texture = texture.NewTexture2DFromRGBA(imageRGBA)

	mat := material.NewStandard(&math32.Color{1, 1, 1})
	mat.AddTexture(s.texture)

	sprite := graphic.NewSprite(s.webcamAspectRatio, 1, mat)

	a.Scene().Add(sprite)
}

func (s *CameraSettings) addGuidelines(a *App) {
	ratio := s.webcamAspectRatio
	geom := geometry.NewGeometry()
	vertices := math32.NewArrayF32(0, 32)
	vertices.Append(
		//crosshairs
		-0.02, 0.0, 0.0,
		0.02, 0.0, 0.0,
		0.0, -0.02, 0.0,
		0.0, 0.02, 0.0,

		//horizontal lines of box
		0.25*ratio, 0.25, 0.0,
		-0.25*ratio, 0.25, 0.0,
		0.25*ratio, -0.25, 0.0,
		-0.25*ratio, -0.25, 0.0,

		//vertical lines of box
		0.25*ratio, 0.25, 0.0,
		0.25*ratio, -0.25, 0.0,
		-0.25*ratio, 0.25, 0.0,
		-0.25*ratio, -0.25, 0.0,
	)
	colors := math32.NewArrayF32(0, 32)
	colors.Append(
		1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
	)
	geom.AddVBO(gls.NewVBO(vertices).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(colors).AddAttrib(gls.VertexColor))

	// Creates basic material
	mat := material.NewBasic()

	// Creates lines with the specified geometry and material
	lines1 := graphic.NewLines(geom, mat)
	a.Scene().Add(lines1)
}

func (s *CameraSettings) getRGBAImageFromWebcam() (*image.RGBA) {
	if ok := s.webcam.Read(&s.mat); !ok {
		fmt.Printf("cannot read device %d\n", s.deviceId)
		return nil
	}

	// mat.ToImg, then type assert as image.RGBA
	// https://stackoverflow.com/questions/31463756/convert-image-image-to-image-nrgba
	img, err := s.mat.ToImage()
	if err != nil {
		fmt.Errorf("Unable to read frame: %v\n", err)
		return nil
	}

	if img, ok := img.(*image.RGBA); ok {
		return img
	} else {
		fmt.Errorf("Unable to convert from gocv.mat to image.RGBA: %v\n")
		return nil
	}
}

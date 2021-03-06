package app

import (
	"fmt"
	"image"
	"os"

	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/math32"
	"gocv.io/x/gocv"
)

type CameraSettings struct {
	devId    *gui.Edit
	deviceId int
	c        chan os.Signal
	mat      gocv.Mat
	window   *gocv.Window
	webcam   *gocv.VideoCapture
	img      *gui.Image
}

func (s *CameraSettings) Initialize(a *App) {
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

	// gocv logic
	// channel to receive os signal
	//s.c = make(chan os.Signal, 1)
	//signal.Notify(s.c, os.Interrupt)
}

func (s *CameraSettings) Render(a *App) {
	if ok := s.webcam.Read(&s.mat); !ok {
		fmt.Printf("cannot read device %d\n", s.deviceId)
		return
	}

	// mat.ToImg, then type assert as image.RGBA
	// https://stackoverflow.com/questions/31463756/convert-image-image-to-image-nrgba
	img, err := s.mat.ToImage()
	if err != nil {
		fmt.Errorf("Unable to read frame: %v\n", err)
		return
	}

	if img, ok := img.(*image.RGBA); ok {
		s.img = gui.NewImageFromRGBA(img)
		a.GuiPanel().Add(s.img)
	}
	//col :=
}

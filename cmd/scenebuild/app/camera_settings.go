package app

import (
	"fmt"
	"image"
	"os"
	"strconv"

	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/texture"
	"gocv.io/x/gocv"
)

type CameraSettings struct {
	devId    *gui.Edit
	deviceId int
	c        chan os.Signal
	mat      gocv.Mat
	window   *gocv.Window
	webcam   *gocv.VideoCapture
	texture  *texture.Texture2D
	app      *App
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

	// Adds control panel after the header
	cpanel := gui.NewPanel(800, 120)
	cpanel.SetBorders(0, 0, 1, 0)
	cpanel.SetPaddings(4, 4, 4, 4)
	cpanel.SetColor(math32.NewColorHex(0xffca6e))
	cpanel.SetLayoutParams(&gui.DockLayoutParams{Edge: gui.DockTop})

	// Add GUI stuff
	l := gui.NewLabel("Camera Device ID (0 - ?)") // Doesn't change camera yet
	l.SetPosition(0, 0)
	l.SetColor(darkTextColor)
	a.Log().Info("Add label")
	cpanel.Add(l)

	s.devId = gui.NewEdit(50, "0")
	s.devId.SetPosition(200, 0)
	s.devId.Subscribe(gui.OnChange, func(name string, ev interface{}) {
		var err error

		oldId, err := strconv.Atoi(s.devId.Text())
		fmt.Println("old id: %v",oldId)
		s.deviceId, err = strconv.Atoi(s.devId.Text())
		if err != nil {
			fmt.Printf("invalid webcam device Id from gui")
			return
		}
		s.webcam, err = gocv.VideoCaptureDevice(s.deviceId)
		if err != nil {
			fmt.Printf("error opening video capture device: %v\n", s.deviceId)
			s.deviceId = oldId
			s.devId.SetText(string(oldId)) 
			s.webcam, err = gocv.VideoCaptureDevice(s.deviceId)
		}
	})
	a.Log().Info("Add dev id")
	cpanel.Add(s.devId)

	a.GuiPanel().Add(cpanel)

	s.makeWebcamView(a)
	//imageRGBA := s.getRGBAImageFromWebcam()
	//image := gui.NewImageFromRGBA(imageRGBA)
	//image.SetPosition(75, 128)
	//s.texture = texture.NewTexture2DFromRGBA(imageRGBA)
	//image.SetTexture(s.texture)
	//a.GuiPanel().Add(image) //FIXME Possible source of memory leak

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
	aspectRatio := width / height

	s.texture = texture.NewTexture2DFromRGBA(imageRGBA)

	mat := material.NewStandard(&math32.Color{1, 1, 1})
	mat.AddTexture(s.texture)

	sprite := graphic.NewSprite(aspectRatio, 1, mat)

	a.Scene().Add(sprite)
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

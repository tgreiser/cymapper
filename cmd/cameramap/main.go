package main

import (
	"fmt"
	"image"
	"image/color"

	"gocv.io/x/gocv"
)

func main() {
	radius := 5
	deviceID := 0

	// open webcam
	webcam, err := gocv.VideoCaptureDevice(int(deviceID))
	if err != nil {
		fmt.Printf("error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// open display window
	window := gocv.NewWindow("CyMapper")
	defer window.Close()

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	gray := gocv.NewMat()
	defer gray.Close()

	// color for the rect when light detected
	blue := color.RGBA{0, 0, 255, 0}

	fmt.Printf("start reading camera device: %v\n", deviceID)
	for {
		if ok := webcam.Read(img); !ok {
			fmt.Printf("cannot read device %d\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		gocv.CvtColor(img, gray, gocv.ColorRGBToGray)
		gocv.GaussianBlur(gray, gray, image.Point{X: radius, Y: radius}, 0, 0, gocv.BorderDefault)

		// detect brightest point
		_, _, _, maxLoc := gocv.MinMaxLoc(gray)

		// draw a rectangle around the bright spot
		gocv.Rectangle(img, image.Rect(maxLoc.X-20, maxLoc.Y-20, maxLoc.X+20, maxLoc.Y+20), blue, 3)

		// show the image in the window, and wait 1 millisecond
		window.IMShow(img)
		window.WaitKey(1)
	}
}

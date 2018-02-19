package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"gocv.io/x/gocv"
)

var deviceID = flag.Int("device-id", 0, "Device ID of your webcam")

func main() {
	flag.Parse()

	// open webcam
	webcam, err := gocv.VideoCaptureDevice(int(*deviceID))
	if err != nil {
		fmt.Printf("error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// open display window
	window := gocv.NewWindow("CyMapper")
	defer window.Close()

	// prepare image matricies
	img := gocv.NewMat()
	defer img.Close()

	// channel to receive os signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	stop := false
	for {
		select {
		case _ = <-c:
			stop = true
		default:
			if ok := webcam.Read(img); !ok {
				fmt.Printf("cannot read device %d\n", *deviceID)
				return
			}

			window.IMShow(img)
			window.WaitKey(1)
		}
		if stop == true {
			break
		}
	}
}

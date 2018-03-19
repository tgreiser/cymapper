package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/tarm/serial"
	"gocv.io/x/gocv"
)

/*
This program is not currently working. The intention is to map 3 pixels at
the same time. One red, blue and green.
*/

var tsvPath = flag.String("file", "output.tsv", "Filename for the tsv output")
var leds = flag.Int("leds", 460, "Number of LEDs per strip (1-10000)")
var pins = flag.Int("pins", 8, "Number of pins which have LEDs connected")
var radius = flag.Int("radius", 7, "Radius of the gaussian blur used for noise reduction")

var deviceID = flag.Int("device-id", 0, "Device ID of your webcam")
var comPort = flag.String("com", "COM8", "COM port for teensy")
var delayMs = flag.Int("delay-ms", 1000, "Number of milliseconds to pause on each LED")
var startPin = flag.Int("start-pin", 1, "Skip to a certain pin")

// Illuminate each LED three at a time, in sequence.
var counter = 0
var max = 0

// Return a buffer of bytes, leds * pins * 3
var bufLen = 0

// color for the rect when light detected
var cb = color.RGBA{0, 0, 255, 0}
var cr = color.RGBA{255, 0, 0, 0}
var cg = color.RGBA{0, 255, 0, 0}

var ticker = time.NewTicker(time.Millisecond * time.Duration(*delayMs))
var stop = false
var width = 0
var height = 0

func init() {
	flag.Parse()

	// ensure radius is above 0 and an odd number
	if *radius < 1 {
		*radius = 1
	}
	if *radius%2 == 0 {
		*radius = *radius + 1
	}

	max = *leds * *pins
	// Return a buffer of bytes, leds * pins * 3
	bufLen = max * 3
	counter = (*startPin - 1) * *leds * 3
}

func main() {
	// Serial configuration for teensy
	c := &serial.Config{Name: *comPort, Baud: 256000}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatalf("When connecting to port: %v: %v", *comPort, err)
	}
	defer s.Close()

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

	red := gocv.NewMat()
	defer red.Close()
	blue := gocv.NewMat()
	defer blue.Close()
	green := gocv.NewMat()
	defer green.Close()

	// read camera dimensions
	if ok := webcam.Read(img); !ok {
		fmt.Printf("cannot read device %d\n", *deviceID)
		return
	}
	fmt.Printf("%d x %d\n", img.Cols(), img.Rows())
	width = img.Cols()
	height = img.Rows()

	file, err := os.Create(*tsvPath)
	if err != nil {
		log.Fatalf("Unable to create %v: %v\n", *tsvPath, err)
	}
	defer file.Close()
	w := csv.NewWriter(file)
	defer w.Flush()
	w.Comma = '\t'

	// channel to receive os signal
	cs := make(chan os.Signal, 1)
	signal.Notify(cs, os.Interrupt)

	// channel to receive camera event
	c1 := make(chan string)
	tick(s, *delayMs, c1)

	fmt.Printf("start reading camera device: %v\n", *deviceID)
	iX := 0
	max := *pins * *leds
	for {
		select {
		case msg := <-c1:
			fmt.Printf("msg: %v\n", msg)
			if ok := webcam.Read(img); !ok {
				fmt.Printf("cannot read device %d\n", *deviceID)
				return
			}
			r, g, b := processFrame(window, img, red, blue, green)
			// stop writing after all the points requested
			write := func(w *csv.Writer, pt *image.Point) {
				if iX >= max {
					return
				}
				err := w.Write([]string{strconv.Itoa(pt.X), strconv.Itoa(pt.Y)})
				if err != nil {
					fmt.Printf("Can not write TSV data: %v\n", err)
				}
				iX++
			}
			write(w, r)
			write(w, g)
			write(w, b)

			if msg == "stop" {
				ticker.Stop()
				stop = true
			}
		case _ = <-cs:
			ticker.Stop()
			stop = true
		}
		if stop == true {
			break
		}
	}
	fmt.Println("Done")
}

func tick(s *serial.Port, delay int, c1 chan string) {
	// start a routine to activate the LEDs
	go func() {
		for _ = range ticker.C {
			ledSequence(s, c1)
			time.AfterFunc(time.Duration(delay/2)*time.Millisecond, func() {
				c1 <- "tick"
			})
		}
	}()
}

func processFrame(window *gocv.Window, img, red, green, blue gocv.Mat) (*image.Point, *image.Point, *image.Point) {
	if img.Empty() {
		return nil, nil, nil
	}
	t := time.Now()

	gocv.GaussianBlur(img, red, image.Point{X: *radius, Y: *radius}, 0, 0, gocv.BorderDefault)

	// split into RGB
	r := img.Rows()
	c := img.Cols() * 3
	red.CopyTo(blue)
	red.CopyTo(green)
	for iX := 0; iX < r; iX++ {
		for iY := 0; iY < c; iY += 3 {
			//fmt.Printf("%v %v\n", iX, iY)
			rb := red.GetSCharAt(iX, iY+2)
			red.SetSCharAt(iX, iY, rb)
			if iY+1 < c {
				red.SetSCharAt(iX, iY+1, rb)
			}

			bb := blue.GetSCharAt(iX, iY)
			if iY+1 < c {
				blue.SetSCharAt(iX, iY+1, bb)
			}
			if iY+2 < c {
				blue.SetSCharAt(iX, iY+2, bb)
			}

			gb := green.GetSCharAt(iX, iY+1)
			green.SetSCharAt(iX, iY+0, gb)
			if iY+2 < c {
				green.SetSCharAt(iX, iY+2, gb)
			}
		}
	}

	gocv.CvtColor(red, red, gocv.ColorRGBToGray)
	gocv.CvtColor(green, green, gocv.ColorRGBToGray)
	gocv.CvtColor(blue, blue, gocv.ColorRGBToGray)

	// detect brightest point
	_, _, _, rLoc := gocv.MinMaxLoc(red)
	_, _, _, gLoc := gocv.MinMaxLoc(green)
	_, _, _, bLoc := gocv.MinMaxLoc(blue)

	// draw a rectangle around the bright spot
	gocv.Rectangle(img, image.Rect(rLoc.X-6, rLoc.Y-6, rLoc.X+6, rLoc.Y+6), cr, 3)
	gocv.Rectangle(img, image.Rect(gLoc.X-6, gLoc.Y-6, gLoc.X+6, gLoc.Y+6), cg, 3)
	gocv.Rectangle(img, image.Rect(bLoc.X-6, bLoc.Y-6, bLoc.X+6, bLoc.Y+6), cb, 3)

	// show the image in the window, and wait 1 millisecond
	window.IMShow(img)
	window.WaitKey(1)
	//window.WaitKey(*delayMs)

	fmt.Printf("R %d x %d\n", rLoc.X, rLoc.Y)
	fmt.Printf("G %d x %d\n", gLoc.X, gLoc.Y)
	fmt.Printf("B %d x %d\n", bLoc.X, bLoc.Y)
	fmt.Printf("%v\n", time.Since(t))
	return &rLoc, &gLoc, &bLoc
}

func ledSequence(s *serial.Port, c chan string) {
	fmt.Printf("Running ledSequence with %d pins, %d LEDs, %d total, %d count\n", *pins, *leds, max, counter)
	buf := make([]byte, bufLen, bufLen)

	for iX := 0; iX < bufLen; iX++ {
		if iX == counter {
			buf[iX] = 0
			buf[iX+1] = 45
			buf[iX+2] = 0

			if iX+5 < bufLen {
				buf[iX+3] = 45
				buf[iX+4] = 0
				buf[iX+5] = 0
			}
			if iX+8 < bufLen {
				buf[iX+6] = 0
				buf[iX+7] = 0
				buf[iX+8] = 45
			}
		}
	}
	counter = counter + 9
	if counter >= bufLen {
		counter = 0
		fmt.Printf("Finished sequence, ending %d\n", bufLen)
		c <- "stop"
	}

	// send to the teensy via serial
	//log.Printf("sending %v bytes\n", len(buf))
	//log.Printf("%v\n", buf)
	_, err := s.Write(buf)
	if err != nil {
		log.Printf("Serial write error: %v\n", err)
	}
}

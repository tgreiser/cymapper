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

var tsvPath = flag.String("file", "output.tsv", "Filename for the tsv output")
var leds = flag.Int("leds", 460, "Number of LEDs per strip (1-10000)")
var pins = flag.Int("pins", 8, "Number of pins which have LEDs connected")
var radius = flag.Int("radius", 7, "Radius of the gaussian blur used for noise reduction")
var brightness = flag.Int("brightness", 64, "LED brightness (1-255)")
var deviceID = flag.Int("device-id", 0, "Device ID of your webcam")
var comPort = flag.String("com", "COM8", "COM port for teensy")
var delayMs = flag.Int("delay-ms", 1000, "Number of milliseconds to pause on each LED")
var startPin = flag.Int("start-pin", 1, "Skip to a certain pin")

// Illuminate each LED one at a time, in sequence.
var counter = 0
var max = 0

// Return a buffer of bytes, leds * pins * 3
var bufLen = 0

// color for the rect when light detected
var blue = color.RGBA{0, 0, 255, 0}

var ticker *time.Ticker
var stop = false
var width = 0
var height = 0

func init() {
	flag.Parse()
	ticker = time.NewTicker(time.Millisecond * time.Duration(*delayMs))

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

	gray := gocv.NewMat()
	defer gray.Close()

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
	tick(s, c1)

	fmt.Printf("start reading camera device: %v with delay %v ms\n", *deviceID, *delayMs)
	for {
		if ok := webcam.Read(img); !ok {
			fmt.Printf("cannot read device %d\n", *deviceID)
			return
		}
		select {
		case msg := <-c1:
			fmt.Printf("%v msg: %v\n", time.Now(), msg)

			go func() {
				pt := processFrame(window, img, gray)
				err := w.Write([]string{strconv.Itoa(pt.X), strconv.Itoa(pt.Y)})
				if err != nil {
					fmt.Printf("Can not write TSV data: %v\n", err)
				}
				window.IMShow(img)
			}()

			if msg == "stop" {
				ticker.Stop()
				stop = true
			}
		case _ = <-cs:
			ticker.Stop()
			stop = true
		default:
			window.IMShow(img)
		}

		window.WaitKey(1)
		if stop == true {
			break
		}
	}
	fmt.Println("Done")
}

func tick(s *serial.Port, c1 chan string) {
	dur := time.Duration(*delayMs/2) * time.Millisecond
	// start a routine to activate the LEDs
	go func() {
		for _ = range ticker.C {
			ledSequence(s, c1)
			fmt.Printf("Take picture in %v ms\n", dur.Seconds()*1000)
			time.AfterFunc(dur, func() {
				c1 <- "tick"
			})
		}
	}()
}

func processFrame(window *gocv.Window, img, gray gocv.Mat) *image.Point {
	if img.Empty() {
		return nil
	}

	gocv.CvtColor(img, gray, gocv.ColorRGBToGray)
	gocv.GaussianBlur(gray, gray, image.Point{X: *radius, Y: *radius}, 0, 0, gocv.BorderDefault)

	// detect brightest point
	_, _, _, maxLoc := gocv.MinMaxLoc(gray)

	// draw a rectangle around the bright spot
	gocv.Rectangle(img, image.Rect(maxLoc.X-6, maxLoc.Y-6, maxLoc.X+6, maxLoc.Y+6), blue, 3)

	// show the image in the window, and wait 1 millisecond
	//window.IMShow(img)
	//window.WaitKey(1)

	fmt.Printf("%d x %d\n", maxLoc.X, maxLoc.Y)
	return &maxLoc
}

func ledSequence(s *serial.Port, c chan string) {
	fmt.Printf("Running ledSequence with %d pins, %d LEDs, %d total, %d count\n", *pins, *leds, max, counter)
	buf := make([]byte, bufLen, bufLen)

	for iX := 0; iX < bufLen; iX++ {
		if iX >= counter && iX < counter+3 {
			buf[iX] = byte(*brightness)
		} else {
			buf[iX] = 0
		}
	}
	counter = counter + 3
	if counter >= bufLen {
		counter = 0
		fmt.Printf("Finished sequence, ending %d\n", bufLen)
		c <- "stop"
	}

	// send to the teensy via serial
	//log.Printf("sending %v bytes\n", len(buf))
	_, err := s.Write(buf)
	if err != nil {
		log.Printf("Serial write error: %v\n", err)
	}
}

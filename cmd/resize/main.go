package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"strconv"
)

var tsvPath = flag.String("file", "remapped.tsv", "Filename for the tsv output")
var vwidth = flag.Int("vwidth", 1280, "Width of the video stream you will be mapping")
var vheight = flag.Int("vheight", 720, "Height of the video stream you will be mapping")
var border = flag.Int("border", 4, "Unused space to leave around the outer pixels")
var flipX = flag.Bool("flip-x", false, "Flip the image along the X axis")
var flipY = flag.Bool("flip-y", true, "Flip the image along the Y axis")

func init() {
	flag.Parse()
}

/**
 * Read TSV data from stdin and then re-map it to the target size
 */
func main() {
	r := csv.NewReader(os.Stdin)
	r.Comma = '\t'
	pts, _ := r.ReadAll()
	fmt.Printf("\ncymapper resize\n")

	p1, p2 := findBounds(pts)
	fmt.Printf("Area with pixels from %d x %d ", p1.X, p1.Y)
	fmt.Printf("to %d x %d\n", p2.X, p2.Y)

	b1, b2 := applyBorder(p1, p2, *border)
	fmt.Printf("Border from %d x %d ", b1.X, b1.Y)
	fmt.Printf("to %d x %d\n", b2.X, b2.Y)

	vsize := image.Point{X: *vwidth, Y: *vheight}

	file, err := os.Create(*tsvPath)
	if err != nil {
		log.Fatalf("Unable to create %v: %v\n", *tsvPath, err)
	}
	defer file.Close()
	w := csv.NewWriter(file)
	defer w.Flush()
	w.Comma = '\t'

	fmt.Printf("Resize %v x %v to %v x %v\n", b2.X-b1.X, b2.Y-b1.Y, *vwidth, *vheight)
	remapPointsAndWrite(pts, b1, b2, vsize, w)
	fmt.Printf("Writing %v\n", *tsvPath)
}

func remapPointsAndWrite(pts [][]string, b1, b2, vsize image.Point, w *csv.Writer) {
	frame := image.Point{X: b2.X - b1.X, Y: b2.Y - b1.Y}
	// after the scene is centered, these represent the basis vectors for the transformation
	xmult := float64(vsize.X) / float64(frame.X)
	if *flipX {
		xmult = xmult * -1
	}
	ymult := float64(vsize.Y) / float64(frame.Y)
	if *flipY {
		ymult = ymult * -1
	}
	fmt.Printf("Transformation: X %v Y %v\n", xmult, ymult)
	for _, pt := range pts {
		ptX, err := strconv.ParseFloat(pt[0], 64)
		if err != nil {
			log.Fatalf("Bad point: %v: %v", pt[0], err)
		}
		ptY, err := strconv.ParseFloat(pt[1], 64)
		if err != nil {
			log.Fatalf("Bad point: %v: %v", pt[1], err)
		}

		lx := (ptX - float64(b1.X)) * xmult
		if *flipX {
			lx += float64(*vwidth)
		}
		ly := (ptY - float64(b1.Y)) * ymult
		if *flipY {
			ly += float64(*vheight)
		}

		w.Write([]string{
			strconv.FormatFloat(lx, 'f', -1, 32),
			strconv.FormatFloat(ly, 'f', -1, 32),
		})
	}
}

func applyBorder(p1, p2 image.Point, border int) (image.Point, image.Point) {
	p1.X -= border
	p1.Y -= border
	p2.X += border
	p2.Y += border
	return p1, p2
}

func findBounds(pts [][]string) (image.Point, image.Point) {
	var p1 = image.Point{X: 9999, Y: 9999}
	var p2 = image.Point{}

	for _, pt := range pts {
		ptX, err := strconv.Atoi(pt[0])
		if err != nil {
			log.Fatalf("Bad point: %v: %v", pt[0], err)
		}
		ptY, err := strconv.Atoi(pt[1])
		if err != nil {
			log.Fatalf("Bad point: %v: %v", pt[1], err)
		}
		if ptX < p1.X {
			p1.X = ptX
		}
		if ptY < p1.Y {
			p1.Y = ptY
		}
		if ptX > p2.X {
			p2.X = ptX
		}
		if ptY > p2.Y {
			p2.Y = ptY
		}
	}

	return p1, p2
}

package fixture

import (
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/g3n/engine/math32"
)

type Fixture struct {
	filepath  string            // File path
	pts       []*math32.Vector3 // List of relative LED coordinates
	tl        math32.Vector3    // Top left corner
	br        math32.Vector3    // Bottom right corner
	idx       int               // internal pointer
	translate math32.Vector3    // translate
	scale     math32.Vector3    // matrix multiply to scale points
}

func New(path string) *Fixture {
	f := new(Fixture)
	f.filepath = path
	tsv, err := os.Open(path)
	if err != nil {
		log.Printf("Invalid TSV file path: %v\n", path)
	}
	reader := csv.NewReader(bufio.NewReader(tsv))
	reader.Comma = '\t'
	var ftlx, ftly float32 = 10000.0, 0.0
	var fbrx, fbry float32 = 0.0, 10000.0
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		x, err := strconv.ParseFloat(line[0], 32)
		if err != nil {
			log.Printf("ERROR: invalid data in %v: %v\n", path, line[0])
			continue
		}
		y, err := strconv.ParseFloat(line[1], 32)
		if err != nil {
			log.Printf("ERROR: invalid data in %v: %v\n", path, line[1])
			continue
		}
		f.pts = append(f.pts, math32.NewVector3(float32(x), float32(y), 0))
		if float32(x) < ftlx {
			ftlx = float32(x)
		}
		if float32(y) > ftly {
			ftly = float32(y)
		}
		if float32(x) > fbrx {
			fbrx = float32(x)
		}
		if float32(y) < fbry {
			fbry = float32(y)
		}
	}
	f.tl = *math32.NewVector3(ftlx, ftly, 0)
	f.br = *math32.NewVector3(fbrx, fbry, 0)
	f.ResetTransformation()
	return f
}

func (f *Fixture) ResetTransformation() {
	f.SetScale(*math32.NewVector3(1.0, 1.0, 1.0))
	f.SetTranslate(*math32.NewVector3(0.0, 0.0, 0.0))
}

func (f *Fixture) Available() bool {
	return f.idx < len(f.pts)
}

func (f *Fixture) Next() *math32.Vector3 {
	defer func() { f.idx++ }()
	return f.pts[f.idx]
}

func (f Fixture) TopLeft() *math32.Vector3 {
	return &f.tl
}

func (f Fixture) BottomRight() *math32.Vector3 {
	return &f.br
}

func (f *Fixture) SetTranslate(t math32.Vector3) {
	f.translate = t
}

func (f *Fixture) SetScale(s math32.Vector3) {
	f.scale = s
}

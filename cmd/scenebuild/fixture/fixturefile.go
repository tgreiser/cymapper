package fixture

import (
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/core"
    "github.com/tgreiser/cymapper/cmd/scenebuild/app"
)

type Fixture struct {
	node      *core.Node        // Main node. It's what you add meshes and 
                                // other nodes to. See the g3nd tank for an example
	filepath  string            // File path
	pts       []*math32.Vector3 // List of relative LED coordinates
	tpts      []*math32.Vector3 // List of transformed coordinates
	tl        *math32.Vector3   // Top left corner
	br        *math32.Vector3   // Bottom right corner
	ttl       *math32.Vector3   // Transformed Top left corner
	tbr       *math32.Vector3   // Transformed Bottom right corner
	idx       int               // internal pointer
	translate *math32.Vector3   // translate
	scale     *math32.Vector3   // matrix multiply to scale points
}

func NewFixture(path string, app *App) {
	f := new(Fixture)
	f.filepath = path
	tsv, err := os.Open(path)
	if err != nil {
		log.Printf("Invalid TSV file path: %v\n", path)
	}
	reader := csv.NewReader(bufio.NewReader(tsv))
	reader.Comma = '\t'
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
	}
	f.tl, f.br = f.FindCorners(f.pts)
	f.ResetTransformation()
    app.fixtures = append(app.fixtures, f)
	return f
}

func (f *Fixture) FindCorners(pts []*math32.Vector3) (topLeft, bottomRight *math32.Vector3) {
	var ftlx, ftly float32 = 10000.0, 0.01
	var fbrx, fbry float32 = 0.0, 10000.0
	for _, p := range pts {
		if float32(p.X) < ftlx {
			ftlx = float32(p.X)
		}
		if float32(p.Y) > ftly {
			ftly = float32(p.Y)
		}
		if float32(p.X) > fbrx {
			fbrx = float32(p.X)
		}
		if float32(p.Y) < fbry {
			fbry = float32(p.Y)
		}
	}
	return math32.NewVector3(ftlx, ftly, 0), math32.NewVector3(fbrx, fbry, 0)
}

func (f *Fixture) ResetTransformation() {
	f.Transform(math32.NewVector3(1.0, 1.0, 1.0),
		math32.NewVector3(0.0, 0.0, 0.0))
}

func (f *Fixture) Available() bool {
	return f.idx < len(f.tpts)
}

func (f *Fixture) Next() *math32.Vector3 {
	defer func() { f.idx++ }()
	return f.tpts[f.idx]
}

func (f *Fixture) Reset() {
	f.idx = 0
}

func (f Fixture) TopLeft() *math32.Vector3 {
	return f.tl
}

func (f Fixture) BottomRight() *math32.Vector3 {
	return f.br
}

func (f Fixture) TransformedTopLeft() *math32.Vector3 {
	return f.ttl
}

func (f Fixture) TransformedBottomRight() *math32.Vector3 {
	return f.tbr
}

func (f *Fixture) Transformed() []*math32.Vector3 {
	f.tpts = make([]*math32.Vector3, len(f.pts), len(f.pts))
	for iP, p := range f.pts {
		f.tpts[iP] = math32.NewVector3(
			(p.X*f.scale.X)+f.translate.X,
			(p.Y*f.scale.Y)+f.translate.Y, 0)
	}
	return f.tpts
}

func (f *Fixture) Transform(scale, translate *math32.Vector3) {
	f.scale = scale
	f.translate = translate
	f.ttl, f.tbr = f.FindCorners(f.Transformed())
}

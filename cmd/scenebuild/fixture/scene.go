package fixture

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
)

type Scene struct {
	fixtures []*Fixture
}

func NewScene(fixtures []*Fixture) *Scene {
	sc := Scene{fixtures}
	return &sc
}

func (s *Scene) SaveAs(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Unable to create %v: %v\n", filename, err)
		return err
	}
	defer file.Close()
	w := csv.NewWriter(file)
	defer w.Flush()
	w.Comma = '\t'

	l := len(s.fixtures)
	for iX := 0; iX < l; iX++ {
		// add fixture vectors to scene
		s.fixtures[iX].Reset()
		for s.fixtures[iX].Available() {
			pt := s.fixtures[iX].Next()
			err := w.Write([]string{strconv.FormatFloat(float64(pt.X), 'f', -1, 64),
				strconv.FormatFloat(float64(pt.Y), 'f', -1, 64)})
			if err != nil {
				log.Printf("%v\n", err)
				return err
			}
		}
	}
	return nil
}

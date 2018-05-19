package fixture

import (
	"testing"

	"github.com/g3n/engine/math32"
)

func TestNewTransformation(t *testing.T) {
	assertTransform(t, 100, 100, 200, 200,
		100, 100, 300, 300,
		2, 2, -100, -100)
	assertTransform(t, 100, 100, 200, 200,
		200, 200, 300, 300,
		1, 1, 100, 100)
	assertTransform(t, 100, 100, 200, 200,
		200, 200, 400, 400,
		2, 2, 0, 0)
	assertTransform(t, 220, 342, 388, 3,
		304, 342, 388, 2,
		.5, 1, 194, 0)
}

func assertTransform(t *testing.T, tlx, tly, brx, bry, ntlx, ntly, nbrx, nbry, scx, scy, trx, try float32) {
	sc, tr := NewTransformation(
		math32.NewVector3(tlx, tly, 0), math32.NewVector3(brx, bry, 0),
		math32.NewVector3(ntlx, ntly, 0), math32.NewVector3(nbrx, nbry, 0),
	)
	if sc.X != scx || sc.Y != scy {
		t.Errorf("Scale was incorrect %v x %v did not match %v\n", scx, scy, sc)
	}
	if tr.X != trx || tr.Y != try {
		t.Errorf("Translate was incorrect %v x %v did not match %v\n", trx, try, tr)
	}
}

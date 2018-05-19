package fixture

import "github.com/g3n/engine/math32"

func NewTransformation(topLeft, bottomRight, newTopLeft, newBottomRight *math32.Vector3) (scale, translate *math32.Vector3) {
	scale = math32.NewVector3((newBottomRight.X-newTopLeft.X)/(bottomRight.X-topLeft.X),
		(newBottomRight.Y-newTopLeft.Y)/(bottomRight.Y-topLeft.Y), 1)
	translate = math32.NewVector3(newTopLeft.X-(scale.X*topLeft.X),
		newTopLeft.Y-(scale.Y*topLeft.Y),
		0)
	return scale, translate
}

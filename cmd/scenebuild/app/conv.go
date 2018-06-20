package app

import (
	"strconv"
)

func ParseFloat32(input string, def float32) float32 {
	out, err := strconv.ParseFloat(input, 32)
	if err != nil {
		return def
	}
	return float32(out)
}

func FormatFloat32(input float32) string {
	return strconv.FormatFloat(float64(input), 'f', -1, 32)
}

func FormatFloat64(input float64) string {
	return strconv.FormatFloat(input, 'f', -1, 64)
}

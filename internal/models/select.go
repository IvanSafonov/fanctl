package models

import (
	"slices"
)

const (
	SelectFuncAverage = "average"
	SelectFuncMax     = "max"
	SelectFuncMin     = "min"
)

var (
	SelectFuncs = []string{SelectFuncAverage, SelectFuncMax, SelectFuncMin}
)

func SelectFunc(selectFunc string) func(values []float64) float64 {
	switch selectFunc {
	case SelectFuncMin:
		return slices.Min
	case SelectFuncAverage:
		return SelectAverage
	default:
		return slices.Max
	}
}

func SelectAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, value := range values {
		sum += value
	}

	return sum / float64(len(values))
}

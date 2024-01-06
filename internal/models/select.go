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

func SelectValue(selectFunc string, values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	switch selectFunc {
	case SelectFuncMin:
		return slices.Min(values)
	case SelectFuncAverage:
		return SelectAverage(values)
	}

	return slices.Max(values)
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

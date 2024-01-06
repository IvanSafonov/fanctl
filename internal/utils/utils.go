package utils

func Ptr[T any](value T) *T {
	return &value
}

func CorrectValue(value float64, factor, add *float64) float64 {
	if factor != nil {
		value *= *factor
	}

	if add != nil {
		value += *add
	}

	return value
}

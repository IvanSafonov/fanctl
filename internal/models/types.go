package models

import "time"

type Seconds float64

func SecondsPtr(sec float64) *Seconds {
	out := Seconds(sec)
	return &out
}

func (d *Seconds) Duration() time.Duration {
	if d == nil {
		return 0
	}

	return time.Millisecond * time.Duration(*d*1000)
}

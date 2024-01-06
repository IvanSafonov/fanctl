package service

import (
	"fmt"
	"time"

	"github.com/IvanSafonov/fanctl/internal/config"
)

type Fan struct {
	conf   config.Fan
	driver FanDriver

	currentLevel config.Level
	delayStart   time.Time
	updated      time.Time
}

func (f *Fan) CheckLevel(value float64, profile string) (config.Level, bool) {
	level, changed := f.Level(value, profile)
	delay := f.currentLevel.DelayDuration()

	if delay != 0 && changed {
		if f.delayStart.IsZero() {
			f.delayStart = time.Now()
			return f.currentLevel, false
		}

		if time.Since(f.delayStart) < delay {
			return f.currentLevel, false
		}
	}

	f.delayStart = time.Time{}

	return level, changed
}

func (f *Fan) Level(value float64, profile string) (config.Level, bool) {
	if f.currentLevel.Contains(value) {
		return f.currentLevel, false
	}

	level := f.conf.FindLevel(value, profile)
	if level.Level == f.currentLevel.Level {
		return f.currentLevel, false
	}

	return level, true
}

func (f *Fan) SetLevel(level config.Level) error {
	err := f.driver.SetLevel(level.Level)
	if err != nil {
		return fmt.Errorf("set fan (%s) level: %w", f.conf.Name, err)
	}

	f.currentLevel = level
	f.delayStart = time.Time{}
	f.updated = time.Now()

	return nil
}

package config

import (
	"fmt"
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Period  *float64
	Repeat  *float64
	Fans    []Fan
	Sensors []Sensor
	Profile *Profile
}

type Fan struct {
	Name string
	Type string

	Sensors []string
	Select  string

	Level    string
	Delay    *float64
	Levels   Levels
	Profiles []ProfileLevels

	Path string
}

func (f *Fan) FindLevel(value float64, profile string) Level {
	level, found := f.profileLevel(value, profile)
	if !found {
		level, found = f.Levels.Find(value)
	}

	if level.Delay == nil {
		level.Delay = f.Delay
	}

	if !found {
		level.Level = f.Level
	}

	return level
}

func (f *Fan) profileLevel(value float64, profile string) (Level, bool) {
	if profile == "" {
		return Level{}, false
	}

	for _, fp := range f.Profiles {
		if fp.Name == profile {
			return fp.Levels.Find(value)
		}
	}

	return Level{}, false
}

type ProfileLevels struct {
	Name   string
	Levels Levels
	Delay  *float64
}

type Levels []Level

func (l Levels) Find(value float64) (Level, bool) {
	for _, level := range l {
		if level.Contains(value) {
			return level, true
		}
	}

	return Level{}, false
}

type Level struct {
	Min   *float64
	Max   *float64
	Level string
	Delay *float64
}

func (l *Level) IsEmpty() bool {
	return l.Level == ""
}

func (l *Level) Contains(value float64) bool {
	if l.IsEmpty() {
		return false
	}

	if l.Min != nil && value < *l.Min {
		return false
	}

	if l.Max != nil && value > *l.Max {
		return false
	}

	return true
}

func (l *Level) DelayDuration() time.Duration {
	return ToDuration(l.Delay)
}

type Sensor struct {
	Name   string
	Type   string
	Factor *float64
	Add    *float64

	Sensor string
	Label  string
	Select string
	Path   string
}

type Profile struct {
	Type string
	Path string
}

func Load(path string) (Config, error) {
	var config Config

	rawYaml, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("read: %w", err)
	}

	err = yaml.Unmarshal(rawYaml, &config)
	if err != nil {
		return config, fmt.Errorf("unmarshal: %w", err)
	}

	if err := validate(&config); err != nil {
		return config, err
	}

	return config, nil
}

func ToDuration(seconds *float64) time.Duration {
	if seconds == nil {
		return 0
	}

	return time.Millisecond * time.Duration(*seconds*1000)
}

package config

import (
	"fmt"
	"os"
	"time"

	"github.com/goccy/go-yaml"

	"github.com/IvanSafonov/fanctl/internal/models"
)

type Config struct {
	Period  *models.Seconds
	Fans    []Fan
	Sensors []Sensor
	Profile *Profile
}

type Fan struct {
	Name string
	Type string

	Sensors []string
	Select  string

	Level     string
	Repeat    *models.Seconds
	Delay     *models.Seconds
	DelayUp   *models.Seconds `yaml:"delayUp"`
	DelayDown *models.Seconds `yaml:"delayDown"`
	Levels    []Level
	Profiles  []ProfileLevels

	Path     string
	RawLevel bool `yaml:"rawLevel"`
}

type ProfileLevels struct {
	Name      string
	Levels    []Level
	Delay     *models.Seconds
	DelayUp   *models.Seconds `yaml:"delayUp"`
	DelayDown *models.Seconds `yaml:"delayDown"`
}

type Level struct {
	Min       *float64
	Max       *float64
	Level     string
	Delay     *models.Seconds
	DelayUp   *models.Seconds `yaml:"delayUp"`
	DelayDown *models.Seconds `yaml:"delayDown"`
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

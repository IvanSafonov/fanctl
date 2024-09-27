package service

import (
	"cmp"
	"fmt"
	"log/slog"
	"time"

	"github.com/IvanSafonov/fanctl/internal/config"
	"github.com/IvanSafonov/fanctl/internal/models"
)

type Fan struct {
	Name string

	driver          FanDriver
	repeat          time.Duration
	defaultLevel    string
	suspendLevel    string
	defaultLevels   Levels
	profileLevels   map[string]Levels
	selectValueFunc func(map[string]float64) float64

	levels  Levels
	updated time.Time
}

func NewFan(driver FanDriver, conf config.Fan) Fan {
	defaults := NewFanDefaults(driver, conf)
	levels := NewLevels(conf.Levels, defaults)
	profileLevels := make(map[string]Levels, len(conf.Profiles))

	for _, profile := range conf.Profiles {
		profileDefaults := defaults.WithProfile(profile)
		profileLevels[profile.Name] = NewLevels(profile.Levels, profileDefaults)
	}

	selectFunc := models.SelectFunc(conf.Select)
	var selectValueFunc func(map[string]float64) float64
	if len(conf.Sensors) == 0 {
		selectValueFunc = func(values map[string]float64) float64 {
			return selectFunc(allValues(values))
		}
	} else {
		selectValueFunc = func(values map[string]float64) float64 {
			return selectFunc(namedValues(values, conf.Sensors))
		}
	}

	return Fan{
		Name:            conf.Name,
		driver:          driver,
		repeat:          defaults.Repeat.Duration(),
		levels:          levels,
		defaultLevel:    defaults.Level,
		suspendLevel:    defaults.SuspendLevel,
		defaultLevels:   levels,
		profileLevels:   profileLevels,
		selectValueFunc: selectValueFunc,
	}
}

// Switches to profile levels set
func (f *Fan) UpdateProfile(profile string) {
	if len(f.profileLevels) == 0 {
		return
	}

	if pl, ok := f.profileLevels[profile]; ok {
		f.levels = pl
	} else {
		f.levels = f.defaultLevels
	}
}

// Updates fan level according to sensors values
// - select sensor value
// - check and update current level
// - update driver level if level is changed or need to repeat
func (f *Fan) UpdateLevel(values map[string]float64) error {
	value := f.selectValueFunc(values)
	if !f.levels.Update(value) && time.Since(f.updated) < f.repeat {
		return nil
	}

	level := f.levels.Level()
	slog.Info("update level", "fan", f.Name, "level", level, "value", value)

	if err := f.driver.SetLevel(level); err != nil {
		return fmt.Errorf("set fan (%s) level: %w", f.Name, err)
	}

	f.updated = time.Now()
	return nil
}

func (f *Fan) SetDefaultLevel() {
	slog.Info("set default level", "fan", f.Name, "level", f.defaultLevel)

	if err := f.driver.SetLevel(f.defaultLevel); err != nil {
		slog.Error("failed to set default level", "error", err)
	}
}

func (f *Fan) SetSuspendLevel() {
	slog.Info("set suspend level", "fan", f.Name, "level", f.suspendLevel)

	if err := f.driver.SetLevel(f.suspendLevel); err != nil {
		slog.Error("failed to set suspend level", "error", err)
	}
}

func allValues(values map[string]float64) []float64 {
	result := make([]float64, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}

	return result
}

func namedValues(values map[string]float64, names []string) []float64 {
	result := make([]float64, 0, len(values))
	for _, name := range names {
		result = append(result, values[name])
	}

	return result
}

type FanDefaults struct {
	Level        string
	SuspendLevel string
	Repeat       models.Seconds
	DelayUp      *models.Seconds
	DelayDown    *models.Seconds
}

func NewFanDefaults(driver FanDriver, conf config.Fan) FanDefaults {
	drvDefaults := driver.Defaults()
	if conf.Repeat != nil {
		drvDefaults.Repeat = *conf.Repeat
	}

	return FanDefaults{
		Level:        cmp.Or(drvDefaults.Level, conf.Level),
		SuspendLevel: cmp.Or(drvDefaults.Level, conf.SuspendLevel),
		Repeat:       drvDefaults.Repeat,
		DelayUp:      cmp.Or(conf.DelayUp, conf.Delay),
		DelayDown:    cmp.Or(conf.DelayDown, conf.Delay),
	}
}

func (fd FanDefaults) WithProfile(conf config.ProfileLevels) FanDefaults {
	fd.DelayUp = cmp.Or(conf.DelayUp, conf.Delay, fd.DelayUp)
	fd.DelayDown = cmp.Or(conf.DelayDown, conf.Delay, fd.DelayDown)
	return fd
}

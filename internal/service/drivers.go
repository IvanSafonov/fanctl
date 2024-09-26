package service

import (
	"github.com/IvanSafonov/fanctl/internal/config"
	"github.com/IvanSafonov/fanctl/internal/drivers"
	"github.com/IvanSafonov/fanctl/internal/models"
)

type FanDriver interface {
	Init() error
	SetLevel(level string) error
	Defaults() drivers.FanDefaults
}

type ProfileDriver interface {
	Init() error
	State() (string, error)
}

type SensorDriver interface {
	Init() error
	Value() (float64, error)
}

//go:generate mockgen -package service -destination ./drivers_mock_test.go . FanDriver,ProfileDriver,SensorDriver

func createProfile(conf *config.Profile) ProfileDriver {
	if conf == nil {
		return nil
	}

	switch conf.Type {
	case models.ProfileTypePlatform:
		return drivers.NewProfilePlatform(*conf)
	}

	return nil
}

func createSensors(confs []config.Sensor) map[string]SensorDriver {
	sensors := make(map[string]SensorDriver, len(confs))

	for _, conf := range confs {
		switch conf.Type {
		case models.SensorTypeHwmon:
			sensors[conf.Name] = drivers.NewSensorHwmon(conf)
		}
	}

	return sensors
}

func createFans(confs []config.Fan) []Fan {
	fans := make([]Fan, 0, len(confs))

	for _, conf := range confs {
		switch conf.Type {
		case models.FanTypeThinkpad:
			driver := drivers.NewFanThinkpad(conf)
			fans = append(fans, NewFan(driver, conf))
		}
	}

	return fans
}

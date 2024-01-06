package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/IvanSafonov/fanctl/internal/config"
	"github.com/IvanSafonov/fanctl/internal/drivers"
	"github.com/IvanSafonov/fanctl/internal/models"
)

type Service struct {
	period time.Duration
	repeat time.Duration

	profileDriver ProfileDriver
	sensorDrivers map[string]SensorDriver
	fans          []Fan

	currentProfile string
	currentValues  map[string]float64
}

type FanDriver interface {
	Init() error
	SetLevel(level string) error
	DefaultLevel() string
}

type ProfileDriver interface {
	Init() error
	State() (string, error)
}

type SensorDriver interface {
	Init() error
	Value() (float64, error)
}

//go:generate mockgen -package service -destination ./service_mock_test.go . FanDriver,ProfileDriver,SensorDriver

func New(conf config.Config) *Service {
	s := Service{
		profileDriver: createProfile(conf.Profile),
		sensorDrivers: createSensors(conf.Sensors),
		fans:          createFans(conf.Fans),

		currentValues: make(map[string]float64, len(conf.Sensors)),
	}

	if conf.Period == nil {
		s.period = time.Second
	} else {
		s.period = config.ToDuration(conf.Period)
	}

	if conf.Repeat == nil {
		s.repeat = time.Minute
	} else {
		s.repeat = config.ToDuration(conf.Repeat)
	}

	return &s
}

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
			if conf.Level == "" {
				conf.Level = driver.DefaultLevel()
			}

			fans = append(fans, Fan{
				conf:   conf,
				driver: driver,
			})
		}
	}

	return fans
}

func (s *Service) Init() error {
	if s.profileDriver != nil {
		err := s.profileDriver.Init()
		if err != nil {
			return fmt.Errorf("profile init: %w", err)
		}
	}

	for name, sensor := range s.sensorDrivers {
		err := sensor.Init()
		if err != nil {
			return fmt.Errorf("sensor (%s) init: %w", name, err)
		}
	}

	for _, fan := range s.fans {
		err := fan.driver.Init()
		if err != nil {
			return fmt.Errorf("fan (%s) init: %w", fan.conf.Name, err)
		}
	}

	return nil
}

func (s *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.period)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.update(ctx); err != nil {
				return err
			}
		}
	}
}

func (s *Service) update(ctx context.Context) error {
	if err := s.updateValues(); err != nil {
		return err
	}

	if err := s.updateProfile(); err != nil {
		return err
	}

	if slog.Default().Enabled(ctx, slog.LevelDebug) {
		fields := make([]any, 0, 2*len(s.currentValues)+2)
		fields = append(fields, "profile", s.currentProfile)
		for name, value := range s.currentValues {
			fields = append(fields, "sensor_"+name, value)
		}

		slog.Debug("State", fields...)
	}

	for i := range s.fans {
		if err := s.updateFan(&s.fans[i]); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) updateValues() error {
	for name, driver := range s.sensorDrivers {
		value, err := driver.Value()
		if err != nil {
			return fmt.Errorf("get sensor (%s) value: %w", name, err)
		}

		s.currentValues[name] = value
	}

	return nil
}

func (s *Service) updateProfile() error {
	if s.profileDriver == nil {
		return nil
	}

	profile, err := s.profileDriver.State()
	if err != nil {
		return fmt.Errorf("get profile: %w", err)
	}

	if s.currentProfile != profile {
		for i := range s.fans {
			s.fans[i].currentLevel = config.Level{}
		}
	}

	s.currentProfile = profile
	return nil
}

func (s *Service) updateFan(fan *Fan) error {
	value := s.fanValue(fan)
	level, changed := fan.CheckLevel(value, s.currentProfile)

	if !changed && time.Since(fan.updated) < s.repeat {
		return nil
	}

	if err := fan.SetLevel(level); err != nil {
		return err
	}

	fields := make([]any, 0, 2*len(s.currentValues)+8)
	fields = append(fields, "fan", fan.conf.Name, "level", level.Level, "profile",
		s.currentProfile, "value", value)
	for name, value := range s.currentValues {
		fields = append(fields, "sensor_"+name, value)
	}

	slog.Info("Update", fields...)

	return nil
}

func (s *Service) fanValue(fan *Fan) float64 {
	var values []float64
	if len(fan.conf.Sensors) == 0 {
		values = make([]float64, 0, len(s.currentValues))
		for _, value := range s.currentValues {
			values = append(values, value)
		}
	} else {
		values = make([]float64, 0, len(fan.conf.Sensors))
		for _, name := range fan.conf.Sensors {
			if value, ok := s.currentValues[name]; ok {
				values = append(values, value)
			}
		}
	}

	return models.SelectValue(fan.conf.Select, values)
}

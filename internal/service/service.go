package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/IvanSafonov/fanctl/internal/config"
)

type Service struct {
	period time.Duration

	profileDriver ProfileDriver
	sensorDrivers map[string]SensorDriver
	fans          []Fan

	profile string
	values  map[string]float64
}

func New(conf config.Config) *Service {
	s := Service{
		period:        time.Second,
		profileDriver: createProfile(conf.Profile),
		sensorDrivers: createSensors(conf.Sensors),
		fans:          createFans(conf.Fans),
		values:        make(map[string]float64, len(conf.Sensors)),
	}

	if conf.Period != nil {
		s.period = conf.Period.Duration()
	}

	return &s
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
			return fmt.Errorf("fan (%s) init: %w", fan.Name, err)
		}
	}

	return nil
}

func (s *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.period)

	for {
		select {
		case <-ctx.Done():
			s.setDefaultLevel()
			return nil
		case <-ticker.C:
			if err := s.Update(ctx); err != nil {
				return err
			}
		}
	}
}

// Updates service state
// - Collect all sensor values to currentValues
// - Update current profile
// - Update fan level
func (s *Service) Update(ctx context.Context) error {
	if err := s.updateValues(); err != nil {
		return err
	}

	if err := s.updateProfile(); err != nil {
		return err
	}

	if slog.Default().Enabled(ctx, slog.LevelDebug) {
		fields := make([]any, 0, 2*len(s.values)+2)
		fields = append(fields, "profile", s.profile)
		for name, value := range s.values {
			fields = append(fields, "sensor_"+name, value)
		}

		slog.Debug("state", fields...)
	}

	for i := range s.fans {
		if err := s.fans[i].UpdateLevel(s.values); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) setDefaultLevel() {
	for i := range s.fans {
		s.fans[i].SetDefaultLevel()
	}
}

func (s *Service) updateValues() error {
	for name, driver := range s.sensorDrivers {
		value, err := driver.Value()
		if err != nil {
			return fmt.Errorf("get sensor (%s) value: %w", name, err)
		}

		s.values[name] = value
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

	if s.profile != profile {
		slog.Info("profile changed", "profile", profile)
		s.profile = profile

		for i := range s.fans {
			s.fans[i].UpdateProfile(profile)
		}
	}

	return nil
}

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/IvanSafonov/fanctl/internal/config"
	"github.com/IvanSafonov/fanctl/internal/drivers"
	"github.com/IvanSafonov/fanctl/internal/models"
	"github.com/IvanSafonov/fanctl/internal/utils"
)

func TestServiceNew(t *testing.T) {
	conf := config.Config{
		Period: models.SecondsPtr(0.2),
		Fans: []config.Fan{
			{Type: models.FanTypeThinkpad, Name: "cpu"},
			{Type: models.FanTypeThinkpad, Name: "gc"},
		},
		Sensors: []config.Sensor{
			{Type: models.SensorTypeHwmon, Name: "cpu"},
			{Type: models.SensorTypeHwmon, Name: "gc"},
		},
		Profile: &config.Profile{
			Type: models.ProfileTypePlatform,
		},
	}

	assert := assert.New(t)

	s := New(conf)

	assert.Equal(200*time.Millisecond, s.period)

	assert.Len(s.fans, 2)
	assert.Equal("cpu", s.fans[0].Name)
	_, ok := s.fans[0].driver.(*drivers.FanThinkpad)
	assert.True(ok)

	assert.Len(s.sensorDrivers, 2)
	_, ok = s.sensorDrivers["gc"].(*drivers.SensorHwmon)
	assert.True(ok)

	_, ok = s.profileDriver.(*drivers.ProfilePlatform)
	assert.True(ok)
}

func TestServiceInit(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)

	fan := NewMockFanDriver(ctrl)
	sensor := NewMockSensorDriver(ctrl)
	profile := NewMockProfileDriver(ctrl)

	s := Service{
		profileDriver: profile,
		sensorDrivers: map[string]SensorDriver{
			"cpu": sensor,
		},
		fans: []Fan{{driver: fan}},
	}

	profile.EXPECT().Init()
	sensor.EXPECT().Init()
	fan.EXPECT().Init()

	err := s.Init()
	assert.NoError(err)
}

func TestServiceRunUnnamedWithProfiles(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)

	fan := NewMockFanDriver(ctrl)
	fan.EXPECT().Defaults().Return(drivers.FanDefaults{Repeat: 1000, Level: "auto"})
	sensor0 := NewMockSensorDriver(ctrl)
	sensor1 := NewMockSensorDriver(ctrl)
	profile := NewMockProfileDriver(ctrl)

	s := New(config.Config{})

	s.period = time.Microsecond
	s.sensorDrivers = map[string]SensorDriver{
		"0": sensor0,
		"1": sensor1,
	}
	s.profileDriver = profile
	s.fans = []Fan{NewFan(
		fan,
		config.Fan{
			Profiles: []config.ProfileLevels{
				{
					Name: "low",
					Levels: []config.Level{
						{Level: "0", Max: utils.Ptr(50.0)},
						{Level: "1", Min: utils.Ptr(45.0), Delay: models.SecondsPtr(1000.0)},
					},
				},
			},
		}),
	}

	ctx, cancel := context.WithCancel(context.Background())

	profile.EXPECT().State().Return("low", nil).AnyTimes()
	sensor0.EXPECT().Value().Return(33.4, nil)
	sensor1.EXPECT().Value().Return(32.6, nil)
	fan.EXPECT().SetLevel("0")

	sensor0.EXPECT().Value().Return(53.4, nil).AnyTimes()
	sensor1.EXPECT().Value().Return(52.6, nil).AnyTimes()
	fan.EXPECT().SetLevel("1").Do(func(level string) {
		cancel()
	})

	fan.EXPECT().SetLevel("auto")

	err := s.Run(ctx)
	assert.NoError(err)
}

func TestServiceRunNamed(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)

	fan := NewMockFanDriver(ctrl)
	fan.EXPECT().Defaults().Return(drivers.FanDefaults{Repeat: 1000, Level: "auto"})
	sensor0 := NewMockSensorDriver(ctrl)
	sensor1 := NewMockSensorDriver(ctrl)

	s := New(config.Config{})

	s.period = time.Microsecond
	s.sensorDrivers = map[string]SensorDriver{
		"nvm": sensor0,
		"cpu": sensor1,
	}
	s.fans = []Fan{NewFan(
		fan,
		config.Fan{
			Sensors: []string{"cpu"},
			Levels: []config.Level{
				{Level: "0", Max: utils.Ptr(50.0)},
				{Level: "1", Min: utils.Ptr(45.0), Delay: models.SecondsPtr(1000.0)},
			},
		}),
	}

	ctx, cancel := context.WithCancel(context.Background())

	sensor0.EXPECT().Value().Return(33.4, nil).Times(2)
	sensor1.EXPECT().Value().Return(32.6, nil).Times(2)
	fan.EXPECT().SetLevel("0")

	sensor0.EXPECT().Value().Return(53.4, nil).AnyTimes()
	sensor1.EXPECT().Value().Return(52.6, nil).AnyTimes()
	fan.EXPECT().SetLevel("1").Do(func(level string) {
		cancel()
	})

	fan.EXPECT().SetLevel("auto")

	err := s.Run(ctx)
	assert.NoError(err)
}

func TestServiceRunRepeat(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)

	fan := NewMockFanDriver(ctrl)
	fan.EXPECT().Defaults().Return(drivers.FanDefaults{Repeat: 0, Level: "auto"})
	sensor0 := NewMockSensorDriver(ctrl)

	s := New(config.Config{})

	s.period = time.Microsecond
	s.sensorDrivers = map[string]SensorDriver{
		"0": sensor0,
	}
	s.fans = []Fan{NewFan(
		fan,
		config.Fan{
			Levels: []config.Level{
				{Level: "0", Max: utils.Ptr(50.0)},
			},
		}),
	}

	ctx, cancel := context.WithCancel(context.Background())

	sensor0.EXPECT().Value().Return(33.4, nil).AnyTimes()
	fan.EXPECT().SetLevel("0").Times(2)

	fan.EXPECT().SetLevel("0").MinTimes(1).Do(func(level string) {
		cancel()
	})

	fan.EXPECT().SetLevel("auto")

	err := s.Run(ctx)
	assert.NoError(err)
}

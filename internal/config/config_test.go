package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/IvanSafonov/fanctl/internal/models"
	"github.com/IvanSafonov/fanctl/internal/utils"
)

func TestConfigLoadFull(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	confFile, err := os.CreateTemp("", "fanctl.yaml")
	require.NoError(err)
	defer os.Remove(confFile.Name())

	_, err = confFile.WriteString(`
    period: 0.5
    repeat: 30
    
    fans:
    - name: cpu
      type: thinkpad
      level: auto
      delay: 3
      select: average
      path: /some/path
      sensors:
      - cpu1
    
      profiles:
      - name: perf
        delay: 4
        levels:
        - min: 10
          max: 55
          level: level 1
          delay: 5
    
      levels:
      - min: 11
        max: 56
        level: level 2
        delay: 6
    
    sensors:
    - name: cpu1
      type: hwmon
      factor: 0.001
      add: 0.5
      select: max
      sensor: coretemp
      label: Core *
    
    profile:
      type: platform
      path: /sys/profile
  `)

	assert.NoError(err)

	config, err := Load(confFile.Name())
	assert.NoError(err)
	assert.Equal(Config{
		Period: utils.Ptr(0.5),
		Repeat: utils.Ptr(30.0),
		Fans: []Fan{
			{
				Name:   "cpu",
				Type:   models.FanTypeThinkpad,
				Level:  "auto",
				Delay:  utils.Ptr(3.0),
				Select: models.SelectFuncAverage,
				Path:   "/some/path",
				Profiles: []ProfileLevels{
					{
						Name:  "perf",
						Delay: utils.Ptr(4.0),
						Levels: []Level{
							{
								Min:   utils.Ptr(10.0),
								Max:   utils.Ptr(55.0),
								Level: "level 1",
								Delay: utils.Ptr(5.0),
							},
						},
					},
				},
				Levels: []Level{
					{
						Min:   utils.Ptr(11.0),
						Max:   utils.Ptr(56.0),
						Level: "level 2",
						Delay: utils.Ptr(6.0),
					},
				},
				Sensors: []string{"cpu1"},
			},
		},
		Sensors: []Sensor{
			{
				Name:   "cpu1",
				Type:   models.SensorTypeHwmon,
				Factor: utils.Ptr(0.001),
				Add:    utils.Ptr(0.5),
				Select: models.SelectFuncMax,
				Sensor: "coretemp",
				Label:  "Core *",
			},
		},
		Profile: &Profile{
			Type: models.ProfileTypePlatform,
			Path: "/sys/profile",
		},
	}, config)
}

func TestLoadConf_ReplacesNotCriticalMistakes(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	confFile, err := os.CreateTemp("", "fanctl.yaml")
	require.NoError(err)
	defer os.Remove(confFile.Name())

	_, err = confFile.WriteString(`
    period: 0.009
    repeat: 0.9
    
    fans:
    - type: thinkpad
      select: fake
      sensors: [s2, s2]
      delay: 101
    
      profiles:
      - name: perf
        delay: -1
        levels:
        - min: 10
          level: 1
          delay: 101
    
      levels:
      - min: 11
        level: 2
        delay: 101
    
    sensors:
    - name: s2
      type: hwmon
      select: fake

    profile:
      type: platform
  `)

	assert.NoError(err)

	config, err := Load(confFile.Name())
	assert.NoError(err)
	assert.Equal(Config{
		Fans: []Fan{
			{
				Name:    "0",
				Type:    models.FanTypeThinkpad,
				Sensors: []string{"s2"},
				Profiles: []ProfileLevels{
					{
						Name: "perf",
						Levels: []Level{
							{
								Min:   utils.Ptr(10.0),
								Level: "1",
							},
						},
					},
				},
				Levels: []Level{
					{
						Min:   utils.Ptr(11.0),
						Level: "2",
					},
				},
			},
		},
		Sensors: []Sensor{
			{
				Type: models.SensorTypeHwmon,
				Name: "s2",
			},
		},
		Profile: &Profile{
			Type: models.ProfileTypePlatform,
		},
	}, config)
}

func TestLoadConfig_Validation(t *testing.T) {
	cases := []struct {
		name string
		err  string
		yml  string
	}{
		{
			name: "empty fans",
			err:  "fans: is empty",
			yml: `
        period: 0.5
        sensors:
        - type: hwmon
      `,
		},
		{
			name: "wrong fan type",
			err:  "fans[0].type: must be one of [thinkpad]",
			yml: `
        sensors:
        - type: hwmon
        fans:
        - type: fake
      `,
		},
		{
			name: "empty fan profile name",
			err:  "fans[0].profiles[0].name: must be set",
			yml: `
        sensors:
        - type: hwmon
        fans:
        - type: thinkpad
          profiles:
          - name: 
      `,
		},
		{
			name: "empty fan profile levels",
			err:  "fans[0]: has no levels",
			yml: `
        sensors:
        - type: hwmon
        fans:
        - type: thinkpad
          profiles:
          - name: perf
            levels: []
      `,
		},
		{
			name: "empty fan level name",
			err:  "fans[0].levels[0].level: must be set",
			yml: `
        sensors:
        - type: hwmon
        fans:
        - type: thinkpad
          levels:
          - min: 1
      `,
		},
		{
			name: "empty fan profile level name",
			err:  "fans[0].profiles[0].levels[0].level: must be set",
			yml: `
        sensors:
        - type: hwmon
        fans:
        - type: thinkpad
          profiles:
          - name: perf
            levels:
            - min: 1
      `,
		},
		{
			name: "fan profile level with no min and max",
			err:  "fans[0].profiles[0].levels[0]: min or max must be set",
			yml: `
        sensors:
        - type: hwmon
        fans:
        - type: thinkpad
          profiles:
          - name: perf
            levels:
            - level: 1
      `,
		},
		{
			name: "using profile without configuration",
			err:  "fans[0].profiles: must set profile configuration",
			yml: `
        sensors:
        - type: hwmon
        fans:
        - type: thinkpad
          profiles:
          - name: perf
            levels:
            - level: 1
              min: 2
      `,
		},
		{
			name: "fan profile level with min greater than max",
			err:  "fans[0].profiles[0].levels[0]: min must be less than max",
			yml: `
        sensors:
        - type: hwmon
        fans:
        - type: thinkpad
          profiles:
          - name: perf
            levels:
            - level: 1
              min: 2.01
              max: 2
      `,
		},
		{
			name: "empty sensors",
			err:  "sensors: is empty",
			yml: `
        fans:
        - type: thinkpad
          levels:
          - level: 1
            max: 2
      `,
		},
		{
			name: "sensors with the same name",
			err:  "sensors[1].name: multiple sensors with the same name",
			yml: `
        fans:
        - type: thinkpad
          levels:
          - level: 1
            max: 2
        sensors:
        - type: hwmon
          name: s1
        - type: hwmon
          name: s1
      `,
		},
		{
			name: "wrong sensor type",
			err:  "sensors[0].type: must be one of [hwmon]",
			yml: `
        fans:
        - type: thinkpad
          levels:
          - level: 1
            max: 2
        sensors:
        - type: fake
      `,
		},
		{
			name: "wrong profile type",
			err:  "profile.type: must be one of [platform]",
			yml: `
        fans:
        - type: thinkpad
          levels:
          - level: 1
            max: 2
        sensors:
        - type: hwmon
        profile:
          type: fake
      `,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			confFile, err := os.CreateTemp("", "fanctl.yaml")
			require.NoError(err)
			defer os.Remove(confFile.Name())

			_, err = confFile.WriteString(tc.yml)
			require.NoError(err)

			_, err = Load(confFile.Name())
			assert.ErrorContains(err, tc.err)
		})
	}
}

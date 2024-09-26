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
    
    fans:
    - name: cpu
      type: thinkpad
      rawLevel: true
      level: auto
      delay: 3
      delayUp: 4
      delayDown: 5
      repeat: 30
      select: average
      path: /some/path
      sensors:
      - cpu1
    
      profiles:
      - name: perf
        delay: 19
        delayUp: 18
        delayDown: 17
        levels:
        - min: 10
          max: 55
          level: level 1
          delay: 5
          delayUp: 6
          delayDown: 7
    
      levels:
      - min: 11
        max: 56
        level: level 2
        delay: 6
        delayUp: 7
        delayDown: 8
    
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
		Period: models.SecondsPtr(0.5),
		Fans: []Fan{
			{
				Name:      "cpu",
				Type:      models.FanTypeThinkpad,
				RawLevel:  true,
				Level:     "auto",
				Delay:     models.SecondsPtr(3.0),
				DelayUp:   models.SecondsPtr(4.0),
				DelayDown: models.SecondsPtr(5.0),
				Repeat:    models.SecondsPtr(30),
				Select:    models.SelectFuncAverage,
				Path:      "/some/path",
				Profiles: []ProfileLevels{
					{
						Name:      "perf",
						Delay:     models.SecondsPtr(19.0),
						DelayUp:   models.SecondsPtr(18.0),
						DelayDown: models.SecondsPtr(17.0),
						Levels: []Level{
							{
								Min:       utils.Ptr(10.0),
								Max:       utils.Ptr(55.0),
								Level:     "level 1",
								Delay:     models.SecondsPtr(5.0),
								DelayUp:   models.SecondsPtr(6.0),
								DelayDown: models.SecondsPtr(7.0),
							},
						},
					},
				},
				Levels: []Level{
					{
						Min:       utils.Ptr(11.0),
						Max:       utils.Ptr(56.0),
						Level:     "level 2",
						Delay:     models.SecondsPtr(6.0),
						DelayUp:   models.SecondsPtr(7.0),
						DelayDown: models.SecondsPtr(8.0),
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
    
    fans:
    - type: thinkpad
      select: fake
      sensors: [s2, s2]
      repeat: 0.9
      delay: 101
      delayUp: 101
      delayDown: 101
    
      profiles:
      - name: perf
        delay: -1
        delayUp: -2
        delayDown: -3
        levels:
        - min: 10
          level: speed 1
          delay: 101
          delayUp: 102
          delayDown: 103
    
      levels:
      - min: 11
        level:   level 7  
        delay: 101
        delayUp: 101
        delayDown: 101
    
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
								Level: "speed 1",
							},
						},
					},
				},
				Levels: []Level{
					{
						Min:   utils.Ptr(11.0),
						Level: "7",
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

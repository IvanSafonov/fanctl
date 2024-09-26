package config

import (
	"cmp"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/IvanSafonov/fanctl/internal/models"
)

func validate(config *Config) error {
	if config.Period != nil && !InRange(0.01, *config.Period, 100) {
		slog.Warn("period: must be within [0.01, 100]")
		config.Period = nil
	}

	if err := validateSensors(config); err != nil {
		return err
	}

	if err := validateProfile(config); err != nil {
		return err
	}

	if err := validateFans(config); err != nil {
		return err
	}

	return nil
}

func validateSensors(config *Config) error {
	if len(config.Sensors) == 0 {
		return errors.New("sensors: is empty")
	}

	names := make(map[string]struct{}, len(config.Sensors))

	for sensorIdx := range config.Sensors {
		sensor := &config.Sensors[sensorIdx]
		sensorPrefix := fmt.Sprintf("sensors[%d]", sensorIdx)

		if sensor.Name != "" {
			if _, exists := names[sensor.Name]; exists {
				return fmt.Errorf("%s.name: multiple sensors with the same name", sensorPrefix)
			}

			names[sensor.Name] = struct{}{}
		} else {
			sensor.Name = strconv.Itoa(sensorIdx)
		}

		if !slices.Contains(models.SensorTypes, sensor.Type) {
			return fmt.Errorf("%s.type: must be one of [%s]", sensorPrefix, strings.Join(models.SensorTypes, ", "))
		}

		if !validateSelect(sensor.Select, sensorPrefix) {
			sensor.Select = ""
		}
	}

	return nil
}

func validateProfile(config *Config) error {
	if config.Profile == nil {
		return nil
	}

	if !slices.Contains(models.ProfileTypes, config.Profile.Type) {
		return fmt.Errorf("profile.type: must be one of [%s]", strings.Join(models.ProfileTypes, ", "))
	}

	return nil
}

func validateFans(config *Config) error {
	if len(config.Fans) == 0 {
		return errors.New("fans: is empty")
	}

	names := make(map[string]struct{}, len(config.Fans))

	for fanIdx := range config.Fans {
		fan := &config.Fans[fanIdx]
		fanPrefix := fmt.Sprintf("fans[%d]", fanIdx)

		if fan.Name != "" {
			if _, exists := names[fan.Name]; exists {
				slog.Warn(fmt.Sprintf("%s.name: multiple fans with the same name", fanPrefix))
			}

			names[fan.Name] = struct{}{}
		} else {
			fan.Name = strconv.Itoa(fanIdx)
		}

		if !slices.Contains(models.FanTypes, fan.Type) {
			return fmt.Errorf("%s.type: must be one of [%s]", fanPrefix, strings.Join(models.FanTypes, ", "))
		}

		if fan.Repeat != nil && !InRange(1, *fan.Repeat, 3600) {
			slog.Warn(fmt.Sprintf("%s.repeat: must be within [1, 3600]", fanPrefix))
			fan.Repeat = nil
		}

		validateDelay(&fan.Delay, fanPrefix+".delay")
		validateDelay(&fan.DelayUp, fanPrefix+".delayUp")
		validateDelay(&fan.DelayDown, fanPrefix+".delayDown")

		if !validateSelect(fan.Select, fanPrefix) {
			fan.Select = ""
		}

		if err := validateLevels(fan.Levels, fanPrefix, fan); err != nil {
			return err
		}

		var levelsCount int

		for profileIdx := range fan.Profiles {
			profile := &fan.Profiles[profileIdx]
			profile.Name = strings.TrimSpace(profile.Name)
			profilePrefix := fmt.Sprintf("%s.profiles[%d]", fanPrefix, profileIdx)

			if profile.Name == "" {
				return fmt.Errorf("%s.name: must be set", profilePrefix)
			}

			validateDelay(&profile.Delay, profilePrefix+".delay")
			validateDelay(&profile.DelayUp, profilePrefix+".delayUp")
			validateDelay(&profile.DelayDown, profilePrefix+".delayDown")

			if err := validateLevels(profile.Levels, profilePrefix, fan); err != nil {
				return err
			}

			levelsCount += len(profile.Levels)
		}

		if levelsCount+len(fan.Levels) == 0 {
			return fmt.Errorf("%s: has no levels", fanPrefix)
		}

		if len(fan.Profiles) != 0 && config.Profile == nil {
			return fmt.Errorf("%s.profiles: must set profile configuration", fanPrefix)
		}

		if len(fan.Sensors) != 0 {
			uniqueSensors := map[string]struct{}{}
			for _, sensor := range fan.Sensors {
				uniqueSensors[sensor] = struct{}{}
			}

			if len(uniqueSensors) != len(fan.Sensors) {
				slog.Warn(fmt.Sprintf("%s.sensors: not unique", fanPrefix))

				fan.Sensors = make([]string, 0, len(uniqueSensors))
				for sensor := range uniqueSensors {
					fan.Sensors = append(fan.Sensors, sensor)
				}
			}

			for _, fanSensor := range fan.Sensors {
				exists := slices.ContainsFunc(config.Sensors, func(sc Sensor) bool {
					return sc.Name == fanSensor
				})

				if !exists {
					return fmt.Errorf("%s.sensors: sensor '%s' not found", fanPrefix, fanSensor)
				}
			}
		}
	}

	return nil
}

func validateLevels(levels []Level, paramPrefix string, fan *Fan) error {
	if len(levels) == 0 {
		return nil
	}

	for levelIdx := range levels {
		level := &levels[levelIdx]
		levelPrefix := fmt.Sprintf("%s.levels[%d]", paramPrefix, levelIdx)

		if level.Level == "" {
			return fmt.Errorf("%s.level: must be set", levelPrefix)
		}
		level.Level = validateLevel(level.Level, levelPrefix, fan)

		if level.Min == nil && level.Max == nil {
			return fmt.Errorf("%s: min or max must be set", levelPrefix)
		}

		if level.Min != nil && level.Max != nil && *level.Min >= *level.Max {
			return fmt.Errorf("%s: min must be less than max", levelPrefix)
		}

		validateDelay(&level.Delay, levelPrefix+".delay")
		validateDelay(&level.DelayUp, levelPrefix+".delayUp")
		validateDelay(&level.DelayDown, levelPrefix+".delayDown")
	}

	return nil
}

func validateDelay(delay **models.Seconds, paramName string) {
	if *delay != nil && !InRange(0, **delay, 100) {
		slog.Warn(fmt.Sprintf("%s: must be within [0, 100]", paramName))
		*delay = nil
	}
}

func validateSelect(value, paramPrefix string) bool {
	if value != "" && !slices.Contains(models.SelectFuncs, value) {
		slog.Warn(fmt.Sprintf("%s.select: must be one of [%s]", paramPrefix, strings.Join(models.SelectFuncs, ", ")))
		return false
	}

	return true
}

func validateLevel(level, paramPrefix string, fan *Fan) string {
	if fan.Type != models.FanTypeThinkpad || fan.RawLevel {
		return level
	}

	level = strings.TrimPrefix(strings.TrimSpace(level), "level ")
	if !slices.Contains(thinkpadLevels, level) {
		slog.Warn(fmt.Sprintf("%s.level: should be one of [%s]",
			paramPrefix, strings.Join(thinkpadLevels, ", ")))
	}
	return level
}

var thinkpadLevels = []string{"0", "1", "2", "3", "4", "5", "6", "7",
	"auto", "disengaged", "full-speed"}

func InRange[T cmp.Ordered](min T, value T, max T) bool {
	return value >= min && value <= max
}

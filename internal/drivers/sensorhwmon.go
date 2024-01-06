package drivers

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/IvanSafonov/fanctl/internal/config"
	"github.com/IvanSafonov/fanctl/internal/models"
	"github.com/IvanSafonov/fanctl/internal/utils"
)

type SensorHwmon struct {
	conf       config.Sensor
	inputFiles []string
}

func NewSensorHwmon(conf config.Sensor) *SensorHwmon {
	if conf.Path == "" {
		conf.Path = "/sys/class/hwmon"
	}

	if conf.Sensor == "" {
		conf.Sensor = "coretemp"
	}

	if conf.Factor == nil {
		conf.Factor = utils.Ptr(0.001)
	}

	return &SensorHwmon{
		conf: conf,
	}
}

func (s *SensorHwmon) Init() error {
	sensorsDirs, err := os.ReadDir(s.conf.Path)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}

	for _, entry := range sensorsDirs {
		sensorDir := path.Join(s.conf.Path, entry.Name())
		sensorNameFile := path.Join(sensorDir, "name")
		if _, err := os.Stat(sensorNameFile); os.IsNotExist(err) {
			continue
		}

		sensorName, err := ReadSysFile(sensorNameFile)
		if err != nil {
			return fmt.Errorf("read sensor name: %w", err)
		}

		if !strings.Contains(string(sensorName), s.conf.Sensor) {
			continue
		}

		sensorFiles, err := os.ReadDir(sensorDir)
		if err != nil {
			return fmt.Errorf("read dir: %w", err)
		}

		for _, sensorFileInfo := range sensorFiles {
			if sensorFileInfo.IsDir() || !strings.HasSuffix(sensorFileInfo.Name(), "_label") {
				continue
			}

			if s.conf.Label != "" {
				label, err := ReadSysFile(path.Join(sensorDir, sensorFileInfo.Name()))
				if err != nil {
					return fmt.Errorf("read label: %w", err)
				}

				if !strings.Contains(string(label), s.conf.Label) {
					continue
				}
			}

			inputFile := path.Join(sensorDir, strings.ReplaceAll(sensorFileInfo.Name(), "_label", "_input"))
			if _, err := os.Stat(inputFile); os.IsNotExist(err) {
				continue
			}

			s.inputFiles = append(s.inputFiles, inputFile)
		}
	}

	if len(s.inputFiles) == 0 {
		return errors.New("input files not found")
	}

	return nil
}

func (s *SensorHwmon) Value() (float64, error) {
	values := make([]float64, 0, len(s.inputFiles))

	for _, inputFile := range s.inputFiles {
		data, err := ReadSysFile(inputFile)
		if err != nil {
			return 0, fmt.Errorf("read input: %w", err)
		}

		value, err := strconv.ParseFloat(data, 64)
		if err != nil {
			return 0, fmt.Errorf("parse input: %w", err)
		}

		value = utils.CorrectValue(value, s.conf.Factor, s.conf.Add)
		values = append(values, value)
	}

	result := models.SelectValue(s.conf.Select, values)
	return result, nil
}

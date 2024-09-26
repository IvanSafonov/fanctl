package drivers

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/IvanSafonov/fanctl/internal/config"
	"github.com/IvanSafonov/fanctl/internal/models"
)

type SensorHwmon struct {
	path       string
	sensor     string
	label      string
	factor     float64
	add        float64
	inputFiles []string
	selectFunc func([]float64) float64
}

func NewSensorHwmon(conf config.Sensor) *SensorHwmon {
	factor := 0.001
	if conf.Factor != nil {
		factor = *conf.Factor
	}

	add := 0.0
	if conf.Add != nil {
		add = *conf.Add
	}

	return &SensorHwmon{
		path:       cmp.Or(conf.Path, "/sys/class/hwmon"),
		sensor:     cmp.Or(conf.Sensor, "coretemp"),
		label:      conf.Label,
		factor:     factor,
		add:        add,
		selectFunc: models.SelectFunc(conf.Select),
	}
}

func (s *SensorHwmon) Init() error {
	sensorsDirs, err := os.ReadDir(s.path)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}

	for _, entry := range sensorsDirs {
		sensorDir := path.Join(s.path, entry.Name())
		sensorNameFile := path.Join(sensorDir, "name")
		if _, err := os.Stat(sensorNameFile); os.IsNotExist(err) {
			continue
		}

		sensorName, err := ReadSysFile(sensorNameFile)
		if err != nil {
			return fmt.Errorf("read sensor name: %w", err)
		}

		if !strings.Contains(string(sensorName), s.sensor) {
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

			if s.label != "" {
				label, err := ReadSysFile(path.Join(sensorDir, sensorFileInfo.Name()))
				if err != nil {
					return fmt.Errorf("read label: %w", err)
				}

				if !strings.Contains(string(label), s.label) {
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

		value = value*s.factor + s.add
		values = append(values, value)
	}

	result := s.selectFunc(values)
	return result, nil
}

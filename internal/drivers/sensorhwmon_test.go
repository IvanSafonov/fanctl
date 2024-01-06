package drivers

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/IvanSafonov/fanctl/internal/config"
	"github.com/IvanSafonov/fanctl/internal/models"
)

func TestSensorHwmon(t *testing.T) {
	assert := assert.New(t)

	tmpDir, err := os.MkdirTemp("", "hwmon")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	createFiles(t, tmpDir, map[string]string{
		"hwmon5/name":        "nvme",
		"hwmon6/name":        "coretemp",
		"hwmon6/temp1_label": "Package id 0",
		"hwmon6/temp1_input": "30200",
		"hwmon6/temp2_label": "Core 0",
		"hwmon6/temp2_input": "35200\n",
		"hwmon6/temp3_label": "Core 4",
		"hwmon6/temp3_input": "34500",
	})

	s := NewSensorHwmon(config.Sensor{
		Path:   tmpDir,
		Select: models.SelectFuncAverage,
		Label:  "Core",
	})
	err = s.Init()
	assert.NoError(err)

	value, err := s.Value()
	assert.NoError(err)
	assert.Equal(34.85, value)
}

func createFiles(t *testing.T, root string, files map[string]string) {
	for name, content := range files {
		dir, fileName := path.Split(name)

		fullDir := path.Join(root, dir)
		err := os.MkdirAll(fullDir, 0777)
		require.NoError(t, err)

		file, err := os.Create(path.Join(fullDir, fileName))
		require.NoError(t, err)

		_, err = file.WriteString(content)
		require.NoError(t, err)
	}
}

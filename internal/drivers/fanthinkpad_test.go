package drivers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/IvanSafonov/fanctl/internal/config"
)

func TestFanThinkpadInit(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	acpiFile, err := os.CreateTemp("", "acpi.fan")
	require.NoError(err)
	defer os.Remove(acpiFile.Name())

	fan := NewFanThinkpad(config.Fan{Path: acpiFile.Name()})
	err = fan.Init()
	assert.NoError(err)
}

func TestFanThinkpadSetLevel(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	acpiFile, err := os.CreateTemp("", "acpi.fan")
	require.NoError(err)
	defer os.Remove(acpiFile.Name())

	fan := NewFanThinkpad(config.Fan{Path: acpiFile.Name()})
	err = fan.SetLevel("level 1")
	assert.NoError(err)

	data, err := os.ReadFile(acpiFile.Name())
	assert.NoError(err)

	assert.Equal("level 1", string(data))
}

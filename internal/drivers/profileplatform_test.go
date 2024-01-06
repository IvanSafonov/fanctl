package drivers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/IvanSafonov/fanctl/internal/config"
)

func TestProfilePlatform(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	profile, err := os.CreateTemp("", "platform_profile")
	require.NoError(err)
	defer os.Remove(profile.Name())

	_, err = profile.WriteString("perf\n")
	require.NoError(err)

	p := NewProfilePlatform(config.Profile{Path: profile.Name()})

	err = p.Init()
	assert.NoError(err)

	state, err := p.State()
	assert.NoError(err)
	assert.Equal("perf", state)
}

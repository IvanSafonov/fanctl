package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/IvanSafonov/fanctl/internal/config"
	"github.com/IvanSafonov/fanctl/internal/utils"
)

func TestFanCheckLevel(t *testing.T) {
	ctrl := gomock.NewController(t)
	driver := NewMockFanDriver(ctrl)
	driver.EXPECT().SetLevel(gomock.Any()).AnyTimes()

	fan := Fan{
		driver: driver,
		conf: config.Fan{
			Delay: utils.Ptr(1000.0),
			Levels: []config.Level{
				{Level: "0", Max: utils.Ptr(40.0), Delay: utils.Ptr(1000.0)},
				{Level: "1", Min: utils.Ptr(30.0), Delay: utils.Ptr(1000.0)},
			},
		},
	}

	assert := assert.New(t)

	// first, changed
	assert.True(fan.updated.IsZero())
	level, changed := fan.CheckLevel(30, "")
	assert.Equal("0", level.Level)
	assert.True(changed)
	assert.NoError(fan.SetLevel(level))
	assert.False(fan.updated.IsZero())

	// same value, not changed
	level, changed = fan.CheckLevel(30, "")
	assert.Equal("0", level.Level)
	assert.True(fan.delayStart.IsZero())
	assert.False(changed)

	// new value, not changed, started delay
	level, changed = fan.CheckLevel(41, "")
	assert.Equal("0", level.Level)
	assert.False(fan.delayStart.IsZero())
	delayStart := fan.delayStart
	assert.False(changed)

	// same value, not changed, still in delay
	level, changed = fan.CheckLevel(41, "")
	assert.Equal("0", level.Level)
	assert.Equal(delayStart, fan.delayStart)
	assert.False(changed)

	// same value, changed after delay
	fan.delayStart = time.Now().Add(-time.Hour)
	level, changed = fan.CheckLevel(41, "")
	assert.Equal("1", level.Level)
	assert.True(fan.delayStart.IsZero())
	assert.True(changed)
}

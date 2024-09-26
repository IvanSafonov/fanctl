package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/IvanSafonov/fanctl/internal/config"
	"github.com/IvanSafonov/fanctl/internal/models"
	"github.com/IvanSafonov/fanctl/internal/utils"
)

func TestLevelsEmpty(t *testing.T) {
	l := NewLevels(nil, FanDefaults{
		Level:     "auto",
		DelayUp:   utils.Ptr(models.Seconds(3.5)),
		DelayDown: utils.Ptr(models.Seconds(6.5)),
	})

	assert.Equal(t, 3500*time.Millisecond, l.items[0].delayUp)
	assert.Equal(t, 6500*time.Millisecond, l.items[0].delayDown)

	assert.Equal(t, "auto", l.Level())
	assert.True(t, l.Update(100))
	assert.Equal(t, "auto", l.Level())
	assert.False(t, l.Update(100))
	assert.Equal(t, "auto", l.Level())
}

func TestLevelsDefaultMinMax(t *testing.T) {
	l := NewLevels([]config.Level{
		{Min: utils.Ptr(30.0), Max: utils.Ptr(50.0), Level: "1"},
	}, FanDefaults{Level: "auto"})

	assert.Equal(t, "auto", l.Level())
	assert.True(t, l.Update(0))
	assert.Equal(t, "auto", l.Level())
	assert.False(t, l.Update(0))
	assert.Equal(t, "auto", l.Level())
	assert.True(t, l.Update(40))
	assert.Equal(t, "1", l.Level())
	assert.True(t, l.Update(100))
	assert.Equal(t, "auto", l.Level())
}

func TestLevelsMultipleUpdates(t *testing.T) {
	l := NewLevels([]config.Level{
		{Min: utils.Ptr(30.0), Max: utils.Ptr(50.0), Level: "1"},
		{Min: utils.Ptr(45.0), Max: nil, Level: "2"},
		{Min: nil, Max: utils.Ptr(40.0), Level: "0"},
	}, FanDefaults{})

	assert.Equal(t, "0", l.Level())

	steps := []struct {
		value   float64
		changed bool
		level   string
	}{
		{value: 0, changed: true, level: "0"},
		{value: 0, changed: false, level: "0"},
		{value: 10, changed: false, level: "0"},
		{value: 39.9, changed: false, level: "0"},
		{value: 40.1, changed: true, level: "1"},
		{value: 30.1, changed: false, level: "1"},
		{value: 29.9, changed: true, level: "0"},
		{value: 50.1, changed: true, level: "2"},
		{value: 20.1, changed: true, level: "0"},
		{value: 1000, changed: true, level: "2"},
		{value: -1000, changed: true, level: "0"},
	}

	for _, step := range steps {
		t.Run(fmt.Sprintf("%0.1f->%s", step.value, step.level), func(t *testing.T) {
			assert.Equal(t, step.changed, l.Update(step.value))
			assert.Equal(t, step.level, l.Level())
		})
	}
}

func TestLevelsWithGaps(t *testing.T) {
	l := NewLevels([]config.Level{
		{Min: nil, Max: utils.Ptr(20.0), Level: "0"},
		{Min: utils.Ptr(30.0), Max: utils.Ptr(40.0), Level: "1"},
		{Min: utils.Ptr(50.0), Max: nil, Level: "2"},
	}, FanDefaults{})

	assert.Equal(t, "0", l.Level())

	steps := []struct {
		value   float64
		changed bool
		level   string
	}{
		{value: 25, changed: true, level: "1"},
		{value: 31, changed: false, level: "1"},
		{value: 29, changed: false, level: "1"},
		{value: 45, changed: true, level: "2"},
		{value: 55, changed: false, level: "2"},
		{value: 49, changed: false, level: "2"},
		{value: 10, changed: true, level: "0"},
		{value: 60, changed: true, level: "2"},
	}

	for _, step := range steps {
		t.Run(fmt.Sprintf("%0.1f->%s", step.value, step.level), func(t *testing.T) {
			assert.Equal(t, step.changed, l.Update(step.value))
			assert.Equal(t, step.level, l.Level())
		})
	}
}

func TestLevelsDelay(t *testing.T) {
	l := NewLevels([]config.Level{
		{Min: nil, Max: utils.Ptr(40.0), Level: "0"},
		{DelayUp: utils.Ptr(models.Seconds(3.5)),
			Min: utils.Ptr(30.0), Max: utils.Ptr(60.0), Level: "1"},
		{Min: utils.Ptr(40.0), Max: nil, Level: "2"},
	}, FanDefaults{
		DelayUp:   utils.Ptr(models.Seconds(1.5)),
		DelayDown: utils.Ptr(models.Seconds(5.5)),
	})

	assert.Equal(t, 3500*time.Millisecond, l.items[1].delayUp)
	assert.Equal(t, 5500*time.Millisecond, l.items[1].delayDown)

	assert.Equal(t, "0", l.Level())

	// First update ignores delay
	assert.True(t, l.Update(41))
	assert.Equal(t, "1", l.Level())

	// delay down
	assert.False(t, l.Update(29))
	assert.Equal(t, "1", l.Level())

	// delay down some time later
	assert.False(t, l.delayStart.IsZero())
	l.delayStart = l.delayStart.Add(-1 * time.Millisecond)
	assert.False(t, l.Update(29))
	assert.Equal(t, "1", l.Level())

	// delay up, ignore previous down delay
	assert.False(t, l.delayStart.IsZero())
	l.delayStart = l.delayStart.Add(-9000 * time.Millisecond)
	assert.False(t, l.Update(61))
	assert.Equal(t, "1", l.Level())

	// delay down 5.5, ignore previous up delay
	assert.False(t, l.delayStart.IsZero())
	l.delayStart = l.delayStart.Add(-9000 * time.Millisecond)
	assert.False(t, l.Update(29))
	assert.Equal(t, "1", l.Level())

	// change after delay
	assert.False(t, l.delayStart.IsZero())
	l.delayStart = l.delayStart.Add(-5500 * time.Millisecond)
	assert.True(t, l.Update(29))
	assert.Equal(t, "0", l.Level())
}

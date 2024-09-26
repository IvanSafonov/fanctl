package service

import (
	"cmp"
	"slices"
	"time"

	"github.com/IvanSafonov/fanctl/internal/config"
)

type Levels struct {
	items       []level
	current     int
	delayStart  time.Time
	isDelayUp   bool
	firstUpdate bool
}

// NewLevels creates
func NewLevels(confLevels []config.Level, defaults FanDefaults) Levels {
	items := make([]level, 0, len(confLevels))

	for _, conf := range confLevels {
		items = append(items, level{
			level:     conf.Level,
			min:       conf.Min,
			max:       conf.Max,
			delayUp:   cmp.Or(conf.DelayUp, conf.Delay, defaults.DelayUp).Duration(),
			delayDown: cmp.Or(conf.DelayDown, conf.Delay, defaults.DelayDown).Duration(),
		})
	}

	slices.SortFunc(items, cmpLevels)

	if len(items) == 0 {
		items = []level{defaultLevel(defaults)}
	}

	if first := items[0]; first.min != nil {
		l := defaultLevel(defaults)
		l.max = first.min
		items = append([]level{l}, items...)
	}

	if last := items[len(items)-1]; last.max != nil {
		l := defaultLevel(defaults)
		l.min = last.max
		items = append(items, l)
	}

	return Levels{
		items:       items,
		firstUpdate: true,
	}
}

// Updates current level according to the value. Returns true if the level is changed.
// If the current level has delay, it will be changed the next time it is called after
// delay period.
func (l *Levels) Update(value float64) bool {
	next := l.current
	for l.items[next].min != nil && value < *l.items[next].min {
		next--
	}

	for l.items[next].max != nil && value > *l.items[next].max {
		next++
	}

	if l.firstUpdate {
		l.firstUpdate = false
		l.current = next
		return true
	}

	if next == l.current {
		l.delayStart = time.Time{}
		return false
	}

	if l.hasDelay(next) {
		return false
	}

	l.current = next
	return true
}

func (l *Levels) hasDelay(next int) bool {
	if next > l.current {
		return l.hasDirectionDelay(l.items[l.current].delayUp, true)
	}

	return l.hasDirectionDelay(l.items[l.current].delayDown, false)
}

func (l *Levels) hasDirectionDelay(delay time.Duration, up bool) bool {
	if delay != 0 {
		if l.delayStart.IsZero() || l.isDelayUp != up {
			l.delayStart = time.Now()
			l.isDelayUp = up
			return true
		} else if time.Since(l.delayStart) < delay {
			return true
		}
	}

	l.delayStart = time.Time{}
	return false
}

// Current fan level
func (l *Levels) Level() string {
	return l.items[l.current].level
}

type level struct {
	min       *float64
	max       *float64
	level     string
	delayUp   time.Duration
	delayDown time.Duration
}

func defaultLevel(defaults FanDefaults) level {
	return level{
		level:     defaults.Level,
		delayUp:   defaults.DelayUp.Duration(),
		delayDown: defaults.DelayDown.Duration(),
	}
}

// min == nil is equal to -∞
// max == nil is equal to +∞
func cmpLevels(a, b level) int {
	if a.min != nil && b.min != nil {
		if n := cmp.Compare(*a.min, *b.min); n != 0 {
			return n
		}
	} else if a.min == nil && b.min != nil {
		return -1
	} else if a.min != nil && b.min == nil {
		return 1
	}

	if a.max != nil && b.max != nil {
		return cmp.Compare(*a.max, *b.max)
	} else if a.max != nil && b.max == nil {
		return -1
	} else if a.max == nil && b.max != nil {
		return 1
	}

	return 0
}

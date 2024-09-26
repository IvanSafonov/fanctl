package drivers

import (
	"cmp"
	"os"

	"github.com/IvanSafonov/fanctl/internal/config"
)

type FanThinkpad struct {
	path   string
	prefix string
}

func NewFanThinkpad(conf config.Fan) *FanThinkpad {
	prefix := "level "
	if conf.RawLevel {
		prefix = ""
	}

	return &FanThinkpad{
		path:   cmp.Or(conf.Path, "/proc/acpi/ibm/fan"),
		prefix: prefix,
	}
}

func (f *FanThinkpad) Init() error {
	file, err := os.Open(f.path)
	if err != nil {
		return err
	}

	file.Close()
	return nil
}

func (f *FanThinkpad) SetLevel(level string) error {
	file, err := os.OpenFile(f.path, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(f.prefix + level)
	if err != nil {
		return err
	}

	return nil
}

func (f *FanThinkpad) Defaults() FanDefaults {
	return FanDefaults{
		Level:  "auto",
		Repeat: 60,
	}
}

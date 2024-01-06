package drivers

import (
	"os"

	"github.com/IvanSafonov/fanctl/internal/config"
)

type FanThinkpad struct {
	conf config.Fan
}

func NewFanThinkpad(conf config.Fan) *FanThinkpad {
	if conf.Path == "" {
		conf.Path = "/proc/acpi/ibm/fan"
	}

	return &FanThinkpad{
		conf: conf,
	}
}

func (f *FanThinkpad) Init() error {
	file, err := os.Open(f.conf.Path)
	if err != nil {
		return err
	}

	file.Close()
	return nil
}

func (f *FanThinkpad) SetLevel(level string) error {
	file, err := os.OpenFile(f.conf.Path, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(level)
	if err != nil {
		return err
	}

	return nil
}

func (f *FanThinkpad) DefaultLevel() string {
	return "level auto"
}

package drivers

import (
	"cmp"
	"errors"
	"os"

	"github.com/IvanSafonov/fanctl/internal/config"
)

type ProfilePlatform struct {
	path string
}

func NewProfilePlatform(conf config.Profile) *ProfilePlatform {
	return &ProfilePlatform{
		path: cmp.Or(conf.Path, "/sys/firmware/acpi/platform_profile"),
	}
}

func (p *ProfilePlatform) Init() error {
	if _, err := os.Stat(p.path); os.IsNotExist(err) {
		return errors.New("file not found: " + p.path)
	}

	return nil
}

func (p *ProfilePlatform) State() (string, error) {
	result, err := ReadSysFile(p.path)
	if err != nil {
		return "", err
	}

	return result, nil
}

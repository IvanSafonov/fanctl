package drivers

import (
	"errors"
	"os"

	"github.com/IvanSafonov/fanctl/internal/config"
)

type ProfilePlatform struct {
	conf config.Profile
}

func NewProfilePlatform(conf config.Profile) *ProfilePlatform {
	if conf.Path == "" {
		conf.Path = "/sys/firmware/acpi/platform_profile"
	}

	return &ProfilePlatform{
		conf: conf,
	}
}

func (p *ProfilePlatform) Init() error {
	if _, err := os.Stat(p.conf.Path); os.IsNotExist(err) {
		return errors.New("file not found: " + p.conf.Path)
	}

	return nil
}

func (p *ProfilePlatform) State() (string, error) {
	result, err := ReadSysFile(p.conf.Path)
	if err != nil {
		return "", err
	}

	return result, nil
}

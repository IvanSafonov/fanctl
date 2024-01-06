package drivers

import (
	"bytes"
	"os"
	"syscall"
)

const sysFileBufferSize = 128

func ReadSysFile(name string) (string, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b := make([]byte, sysFileBufferSize)
	n, err := syscall.Read(int(f.Fd()), b)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(b[:n])), nil
}

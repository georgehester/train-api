package secret

import (
	"bytes"
	"errors"
	"os"
)

func LoadSecret(name string) ([]byte, error) {
	data, err := os.ReadFile("/run/secrets/" + name)
	if err != nil {
		return nil, errors.New("Failed To Load Secret")
	}

	return bytes.TrimSpace(data), nil
}

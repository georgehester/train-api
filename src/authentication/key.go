package authentication

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateRandomApplicationKey() (string, error) {
	buffer := make([]byte, 16)

	if _, randomError := rand.Read(buffer); randomError != nil {
		return "", randomError
	}

	return hex.EncodeToString(buffer), nil
}

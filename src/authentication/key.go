package authentication

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
)

func GenerateRandomApplicationKey() (string, error) {
	buffer := make([]byte, 16)

	if _, randomError := rand.Read(buffer); randomError != nil {
		return "", randomError
	}

	return hex.EncodeToString(buffer), nil
}

func GenerateRandomPassword(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("Invalid Password Length")
	}

	const characterSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"
	password := make([]byte, length)
	maximumIndex := big.NewInt(int64(len(characterSet)))

	for index := 0; index < length; index++ {
		randomIndex, randomError := rand.Int(rand.Reader, maximumIndex)
		if randomError != nil {
			return "", randomError
		}

		password[index] = characterSet[randomIndex.Int64()]
	}

	return string(password), nil
}

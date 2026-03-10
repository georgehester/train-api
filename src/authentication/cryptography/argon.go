package cryptography

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type ArgonParameters struct {
	Memory     uint32
	Iterations uint32
	Threads    uint8
	SaltLength uint32
	KeyLength  uint32
}

var parameters = &ArgonParameters{
	Memory:     64 * 1024, // 64MB
	Iterations: 3,
	Threads:    2,
	SaltLength: 16,
	KeyLength:  32,
}

func decodeHash(encoded string) (*ArgonParameters, []byte, []byte, error) {
	encodedSplit := strings.Split(encoded, "$")
	if len(encodedSplit) != 6 {
		return nil, nil, nil, errors.New("Malformed Hash")
	}

	var version int
	if _, scanError := fmt.Sscanf(encodedSplit[2], "v=%d", &version); scanError != nil {
		return nil, nil, nil, scanError
	}
	if version != argon2.Version {
		return nil, nil, nil, errors.New("Invalid Hash Version")
	}

	parameters := &ArgonParameters{}
	if _, scanError := fmt.Sscanf(
		encodedSplit[3], "m=%d,t=%d,p=%d",
		&parameters.Memory, &parameters.Iterations, &parameters.Threads,
	); scanError != nil {
		return nil, nil, nil, scanError
	}

	salt, decodeError := base64.RawStdEncoding.DecodeString(encodedSplit[4])
	if decodeError != nil {
		return nil, nil, nil, decodeError
	}

	hash, decodeError := base64.RawStdEncoding.DecodeString(encodedSplit[5])
	if decodeError != nil {
		return nil, nil, nil, decodeError
	}

	parameters.SaltLength = uint32(len(salt))
	parameters.KeyLength = uint32(len(hash))

	return parameters, salt, hash, nil
}

func Hash(password string) (string, error) {
	salt := make([]byte, parameters.SaltLength)
	if _, saltError := rand.Read(salt); saltError != nil {
		return "", errors.New("Failed To Generate Salt")
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		parameters.Iterations,
		parameters.Memory,
		parameters.Threads,
		parameters.KeyLength,
	)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		parameters.Memory,
		parameters.Iterations,
		parameters.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

func Verify(password string, encoded string) (bool, error) {
	parameters, salt, hash, decodeError := decodeHash(encoded)
	if decodeError != nil {
		return false, decodeError
	}

	comparison := argon2.IDKey(
		[]byte(password),
		salt,
		parameters.Iterations,
		parameters.Memory,
		parameters.Threads,
		parameters.KeyLength,
	)

	if subtle.ConstantTimeCompare(hash, comparison) == 1 {
		return true, nil
	}
	return false, nil
}

package authentication

import (
	"strings"
	"testing"
)

func TestGenerateRandomApplicationKey(t *testing.T) {
	key, err := GenerateRandomApplicationKey()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(key) != 32 {
		t.Fatalf("expected key length 32, got %d", len(key))
	}
}

func TestGenerateRandomPassword(t *testing.T) {
	const length = 24

	password, err := GenerateRandomPassword(length)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(password) != length {
		t.Fatalf("expected password length %d, got %d", length, len(password))
	}

	const characterSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"
	for _, character := range password {
		if !strings.ContainsRune(characterSet, character) {
			t.Fatalf("unexpected character in password: %q", character)
		}
	}
}

func TestGenerateRandomPasswordInvalidLength(t *testing.T) {
	_, err := GenerateRandomPassword(0)
	if err == nil {
		t.Fatalf("expected error for invalid password length")
	}
}

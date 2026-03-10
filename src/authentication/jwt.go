package authentication

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"
	"vulpz/train-api/src/secret"

	"crypto/ed25519"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/ssh"
)

type KeyManager struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

type UserType string

const (
	UserTypeCustomer      UserType = "customer"
	UserTypeAdministrator UserType = "administrator"
)

type Claims struct {
	Type     UserType `json:"type"`
	Id       string   `json:"id"`
	Email    string   `json:"email"`
	Forename string   `json:"forename"`
	Surname  string   `json:"surname"`
	jwt.RegisteredClaims
}

func NewKeyManager() (*KeyManager, error) {
	privateKey, keyError := loadPrivateKey()
	if keyError != nil {
		return nil, keyError
	}

	publicKey, keyError := loadPublicKey()
	if keyError != nil {
		return nil, keyError
	}

	return &KeyManager{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func (keyManager *KeyManager) PrivateKey() ed25519.PrivateKey {
	return keyManager.privateKey
}

func (keyManager *KeyManager) PublicKey() ed25519.PublicKey {
	return keyManager.publicKey
}

func loadPrivateKey() (ed25519.PrivateKey, error) {
	encoded, err := secret.LoadSecret("key-private")
	if err != nil {
		return nil, errors.New("Private Key Could Not Be Loaded")
	}

	bytes, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		return nil, errors.New("Private Key Could Not Be Decoded")
	}

	raw, err := ssh.ParseRawPrivateKey(bytes)
	if err != nil {
		return nil, fmt.Errorf("Private Key Could Not Be Parsed")
	}

	key, ok := raw.(*ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("Private Key Invalid")
	}

	return *key, nil
}

func loadPublicKey() (ed25519.PublicKey, error) {
	encoded, err := secret.LoadSecret("key-public")
	if err != nil {
		return nil, errors.New("Public Key Could Not Be Loaded")
	}

	bytes, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		return nil, errors.New("Public Key Could Not Be Decoded")
	}

	parsed, _, _, _, err := ssh.ParseAuthorizedKey(bytes)
	if err != nil {
		return nil, fmt.Errorf("Public Key Could Not Be Parsed")
	}

	public, ok := parsed.(ssh.CryptoPublicKey)
	if !ok {
		return nil, errors.New("Public Key Invalid")
	}

	ed25519Public, ok := public.CryptoPublicKey().(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("Public Key Invalid Type")
	}

	return ed25519Public, nil
}

func (keyManager *KeyManager) Sign(userType UserType, id string, email string, forename string, surname string) (string, error) {
	if userType != UserTypeCustomer && userType != UserTypeAdministrator {
		return "", fmt.Errorf("invalid user type: %s", userType)
	}

	claims := Claims{
		Type:     userType,
		Id:       id,
		Email:    email,
		Forename: forename,
		Surname:  surname,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(keyManager.privateKey)
}

func (keyManager *KeyManager) Verify(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
				return nil, errors.New("Failed To Decode Token")
			}
			return keyManager.publicKey, nil
		},
	)
	if err != nil {
		return nil, errors.New("Token Invalid")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("Token Claims Invalid")
	}

	if claims.Type != UserTypeCustomer && claims.Type != UserTypeAdministrator {
		return nil, errors.New("User Type Invalid")
	}

	return claims, nil
}

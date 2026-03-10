package handlers

import (
	"vulpz/train-api/src/authentication"
	"vulpz/train-api/src/email"

	"github.com/jackc/pgx/v5"
	"github.com/patrickmn/go-cache"
)

type Environment struct {
	Database    *pgx.Conn
	Cache       *cache.Cache
	KeyManager  *authentication.KeyManager
	EmailClient *email.EmailClient
}

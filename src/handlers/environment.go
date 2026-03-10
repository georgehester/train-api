package handlers

import (
	"vulpz/train-api/src/authentication"

	"github.com/jackc/pgx/v5"
	"github.com/patrickmn/go-cache"
)

type Environment struct {
	Database   *pgx.Conn
	Cache      *cache.Cache
	KeyManager *authentication.KeyManager
}

package handlers

import (
	"vulpz/train-api/src/authentication"
	"vulpz/train-api/src/email"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/patrickmn/go-cache"
)

type Environment struct {
	Database         *pgxpool.Pool
	Cache            *cache.Cache
	KeyManager       *authentication.KeyManager
	EmailClient      *email.EmailClient
	OpenRouterAPIKey string
}

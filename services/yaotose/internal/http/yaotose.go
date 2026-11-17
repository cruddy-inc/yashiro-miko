package http

import (
	db "yaotose/internal/db/sqlc"

	"github.com/redis/go-redis/v9"
)

type Yaotose struct {
	SecretKey string
	Redis     *redis.Client
	Queries   *db.Queries
}

func New(secretKey string, redis *redis.Client, queries *db.Queries) *Yaotose {
	return &Yaotose{
		SecretKey: secretKey,
		Redis:     redis,
		Queries:   queries,
	}
}

package httpy

import (
	db "yaotose/internal/db/sqlc"
	"yaotose/internal/rules"

	"github.com/redis/go-redis/v9"
)

type Yaotose struct {
	SecretKey string
	Redis     *redis.Client
	Queries   *db.Queries
	Engine    *rules.Engine
}

func New(secretKey string, redis *redis.Client, queries *db.Queries, engine *rules.Engine) *Yaotose {

	return &Yaotose{
		SecretKey: secretKey,
		Redis:     redis,
		Queries:   queries,
		Engine:    engine,
	}
}

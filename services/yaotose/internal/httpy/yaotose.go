package httpy

import (
	"tsubaki/domain"
	db "yaotose/internal/db/sqlc"
	"yaotose/internal/rules"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Yaotose struct {
	SecretKey *domain.RerollStore
	Redis     *redis.Client
	Queries   *db.Queries
	Pool      *pgxpool.Pool
	Engine    *rules.Engine
	Kafka     *kgo.Client
	Currency  *domain.CurrencyStore
}

func New(secretKey *domain.RerollStore, redis *redis.Client, queries *db.Queries, pool *pgxpool.Pool, engine *rules.Engine, kafka *kgo.Client, currency *domain.CurrencyStore) *Yaotose {

	return &Yaotose{
		SecretKey: secretKey,
		Redis:     redis,
		Queries:   queries,
		Pool:      pool,
		Engine:    engine,
		Kafka:     kafka,
		Currency:  currency,
	}
}

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"pkg/errlog"
	"pkg/failure"
	"pkg/logger"
	"time"
	"tsubaki/domain"
	db "yaotose/internal/db/sqlc"
	"yaotose/internal/httpy"
	"yaotose/internal/rules"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
)

func main() {
	os.Exit(start())
}

// start first with postgres, redis, go-cel, kafka in finale
func start() int {
	ctx := context.Background()
	l := logger.NewLogger("yaotose")
	log := l.With(slog.String("op", "yaotose"))
	ctx = logger.SetLogger(ctx, log)

	tracer := otel.Tracer("yaotose_main")
	rootCtx, rootSpan := tracer.Start(ctx, "yaotose.main.start")
	defer rootSpan.End()

	envCtx, envSpan := tracer.Start(rootCtx, "yaotose.main.environmental.variable.fetch")
	defer envSpan.End()
	// Use os.Getenv (final version might use viper)
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		err := failure.EmptyValue("DATABASE_URL")
		err = failure.Wrap("yaotose.main.environmental.variable.database.url.fetch", err)
		fmt.Fprintf(os.Stderr, "yaotose.main: %v\n", err)
		errlog.Internal(envCtx, nil, "DATABASE_URL", failure.InternalEmptyValueError, err)
		return 1
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		err := failure.EmptyValue("SECRET_KEY")
		err = failure.Wrap("yaotose.main.environmental.variable.secret.key.fetch", err)
		fmt.Fprintf(os.Stderr, "yaotose.main: %v\n", err)
		errlog.Internal(envCtx, nil, "DATABASE_URL", failure.InternalEmptyValueError, err)
		return 1
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		err := failure.EmptyValue("REDIS_ADDR")
		err = failure.Wrap("yaotose.main.environmental.variable.redis.address.fetch", err)
		fmt.Fprintf(os.Stderr, "yaotose.main: %v\n", err)
		errlog.Internal(envCtx, nil, "REDIS_ADDR", failure.InternalEmptyValueError, err)
		return 1
	}

	redisPass := os.Getenv("REDIS_PASSWORD")
	if redisAddr == "" {
		err := failure.EmptyValue("REDIS_PASSWORD")
		err = failure.Wrap("yaotose.main.environmental.variable.redis.password.fetch", err)
		fmt.Fprintf(os.Stderr, "yaotose.main: %v\n", err)
		errlog.Internal(envCtx, nil, "REDIS_PASSWORD", failure.InternalEmptyValueError, err)
		return 1
	}

	kafkaAddr := os.Getenv("KAFKA_ADDR")
	if kafkaAddr == "" {
		err := failure.EmptyValue("KAFKA_ADDR")
		err = failure.Wrap("yaotose.main.environmental.variable.kafka.address.fetch", err)
		fmt.Fprintf(os.Stderr, "yaotose.main: %v\n", err)
		errlog.Internal(envCtx, nil, "KAFKA_ADDR", failure.InternalEmptyValueError, err)
		return 1
	}

	poolCtx, poolSpan := tracer.Start(rootCtx, "yaotose.main.pool.connection.create")
	defer poolSpan.End()
	conf, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		err = failure.Wrap("yaotose.main.pool.connection.parse.url", err)
		fmt.Fprintf(os.Stderr, "yaotose.main: %v\n", err)
		errlog.Internal(poolCtx, nil, "database_url_parse", failure.InternalEmptyValueError, err)
		return 1
	}

	pool, err := pgxpool.NewWithConfig(poolCtx, conf)
	if err != nil {
		err = failure.Wrap("yaotose.main.pool.connection.create.from.config", err)
		fmt.Fprintf(os.Stderr, "yaotose.main: %v\n", err)
		errlog.Internal(poolCtx, nil, "database_pool_create", failure.InternalGeneralError, err)
		return 1
	}
	defer pool.Close()

	queries := db.New(pool)

	decisionCtx, decisionSpan := tracer.Start(rootCtx, "yaotose.main.decision.part.create")
	defer decisionSpan.End()
	goCelEng, err := rules.NewEngine()
	if err != nil {
		err = failure.Wrap("yaotose.main.cel.engine.create", err)
		fmt.Fprintf(os.Stderr, "yaotose.main: %v\n", err)
		errlog.General(decisionCtx, nil, "go-cel_engine_create", failure.InternalGeneralError, err)
		return 1
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:            redisAddr,
		Password:        redisPass,
		MaxRetries:      1,
		MinRetryBackoff: 1 * time.Millisecond,
		MaxRetryBackoff: 2 * time.Millisecond,

		MinIdleConns:     5,
		PipelinePoolSize: 30,
		MaxIdleConns:     10,

		DialTimeout:  5 * time.Second,
		ReadTimeout:  8 * time.Millisecond,
		WriteTimeout: 8 * time.Millisecond,
		PoolTimeout:  3 * time.Millisecond,
	})

	kafkaClient, err := kgo.NewClient(
		kgo.SeedBrokers(kafkaAddr),
	)
	if err != nil {
		err = failure.Wrap("yaotose.main.kafka.client.create", err)
		fmt.Fprint(os.Stderr, "yaotose.main: %v\n", err)
		errlog.General(decisionCtx, nil, "kafka_client_create", failure.InternalGeneralError, err)
		return 1
	}

	ya := httpy.New(&domain.RerollStore{}, redisClient, queries, pool, goCelEng, kafkaClient, &domain.CurrencyStore{})

	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Route("/yaotose", func(r chi.Router) {
			r.Post("/incoming", ya.IncomingTransaction)
		})
	})

	return 0
}

package tsubaki

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"pkg/errlog"
	"pkg/failure"
	"pkg/logger"
	"syscall"
	"time"
	"tsubaki/domain"
	db "tsubaki/internal/db/sqlc"
	"tsubaki/internal/jobs/currency"
	"tsubaki/internal/jobs/kafka"
	"tsubaki/internal/jobs/reroll"

	pb "tsubaki/internal/gen/miko/v1"
	tsubakigrpc "tsubaki/internal/grpc"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
)

func main() {
	os.Exit(start())
}

func start() int {
	ctx := context.Background()
	l := logger.NewLogger("tsubaki")
	log := l.With(slog.String("op", "tsubaki"))
	ctx = logger.SetLogger(ctx, log)

	tracer := otel.Tracer("tsubaki_main")
	rootCtx, rootSpan := tracer.Start(ctx, "tsubaki.main.start")
	defer rootSpan.End()

	httpClient := http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   20,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	envCtx, envSpan := tracer.Start(rootCtx, "tsubaki.main.environmental.variable.fetch")
	defer envSpan.End()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		err := failure.EmptyValue("DATABASE_URL")
		err = failure.Wrap("tsubaki.main.environmental.variable.database.url.fetch", err)
		fmt.Fprintf(os.Stderr, "tsubaki.main: %v\n", err)
		errlog.Internal(envCtx, nil, "DATABASE_URL", failure.InternalEmptyValueError, err)
		return 1
	}

	kafkaAddr := os.Getenv("KAFKA_ADDR")
	if kafkaAddr == "" {
		err := failure.EmptyValue("KAFKA_ADDR")
		err = failure.Wrap("tsubaki.main.environmental.variable.kafka.addr.fetch", err)
		fmt.Fprintf(os.Stderr, "tsubaki.main: %v\n", err)
		errlog.Internal(envCtx, nil, "KAFKA_ADDR", failure.InternalEmptyValueError, err)
		return 1
	}

	gRPCAddr := os.Getenv("GRPC_ADDR")
	if gRPCAddr == "" {
		err := failure.EmptyValue("GRPC_ADDR")
		err = failure.Wrap("tsubaki.main.environmental.variable.grpc.addr.fetch", err)
		fmt.Fprintf(os.Stderr, "tsubaki.main: %v\n", err)
		errlog.Internal(envCtx, nil, "GRPC_ADDR", failure.InternalEmptyValueError, err)
		return 1
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		err := failure.EmptyValue("SECRET_KEY")
		err = failure.Wrap("tsubaki.main.environmental.variable.secret.key.fetch", err)
		fmt.Fprintf(os.Stderr, "tsubaki.main: %v\n", err)
		errlog.Internal(envCtx, nil, "SECRET_KEY", failure.InternalEmptyValueError, err)
		return 1
	}

	poolCtx, poolSpan := tracer.Start(rootCtx, "tsubaki.main.pool.connection.create")
	defer poolSpan.End()

	conf, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		err = failure.Wrap("tsubaki.main.pool.connection.parse.url", err)
		fmt.Fprintf(os.Stderr, "tsubaki.main: %v\n", err)
		errlog.Internal(poolCtx, nil, "database_url_parse", failure.InternalEmptyValueError, err)
		return 1
	}

	pool, err := pgxpool.NewWithConfig(poolCtx, conf)
	if err != nil {
		err = failure.Wrap("tsubaki.main.pool.connection.create.from.config", err)
		fmt.Fprintf(os.Stderr, "tsubaki.main: %v\n", err)
		errlog.Internal(poolCtx, nil, "database_pool_create", failure.InternalGeneralError, err)
		return 1
	}
	defer pool.Close()

	queries := db.New(pool)

	kafkaCtx, kafkaSpan := tracer.Start(rootCtx, "tsubaki.main.kafka.client.create")
	defer kafkaSpan.End()
	kafkaClient, err := kgo.NewClient(
		kgo.SeedBrokers(kafkaAddr),
	)
	if err != nil {
		err = failure.Wrap("tsubaki.main.kafka.client.create", err)
		fmt.Fprintf(os.Stderr, "tsubaki.main: %v\n", err)
		errlog.General(kafkaCtx, nil, "kafka_client_create", failure.InternalGeneralError, err)
		return 1
	}

	riverCtx, riverSpan := tracer.Start(rootCtx, "tsubaki.main.river.setup")
	defer riverSpan.End()

	riverSetupCtx, riverSetupSpan := tracer.Start(riverCtx, "tsubaki.main.river.setup.create")
	defer riverSetupSpan.End()
	migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		err = failure.Wrap("tsubaki.main.river.migrator.create", err)
		fmt.Fprintf(os.Stderr, "tsubaki.main: %v\n", err)
		errlog.Internal(riverSetupCtx, nil, "river_migrator_create", failure.InternalGeneralError, err)
		return 1
	}

	_, err = migrator.Migrate(context.Background(), rivermigrate.DirectionUp, nil)
	if err != nil {
		err = failure.Wrap("tsubaki.main.river.migrator.migrate", err)
		fmt.Fprintf(os.Stderr, "tsubaki.main: %v\n", err)
		errlog.Internal(riverSetupCtx, nil, "river_migration_failed", failure.InternalGeneralError, err)
		return 1
	}

	periodicJobs := []*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(720*time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return reroll.KeyRerollArgs{}, nil
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(200*time.Millisecond),
			func() (river.JobArgs, *river.InsertOpts) {
				return kafka.CasesOpenedOutboxArgs{}, nil
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(6*time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return currency.CurrencyTrackArgs{}, nil
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
	}

	workers := river.NewWorkers()
	keyRerollWorker := &reroll.KeyRerollWorker{
		Secret: &domain.RerollStore{},
	}
	casesOpenedOutboxWorker := &kafka.CasesOpenedOutboxWorker{
		Queries: queries,
		Kafka:   kafkaClient,
	}
	currencyWorker := &currency.CurrencyTrackWorker{
		HttpClient: &httpClient,
		Rates:      &domain.CurrencyStore{},
	}

	river.AddWorker(workers, keyRerollWorker)
	river.AddWorker(workers, casesOpenedOutboxWorker)
	river.AddWorker(workers, currencyWorker)

	riverRunCtx, riverRunSpan := tracer.Start(riverCtx, "tsubaki.main.river.run")
	defer riverRunSpan.End()
	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Workers: workers,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {
				MaxWorkers: 15,
			},
		},
		PeriodicJobs: periodicJobs,
	})
	if err != nil {
		err = failure.Wrap("tsubaki.main.river.client.create", err)
		fmt.Fprintf(os.Stderr, "tsubaki.main: %v\n", err)
		errlog.Internal(riverRunCtx, nil, "river_client_create", failure.InternalGeneralError, err)
		return 1
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err = riverClient.Start(ctx)
	if err != nil {
		err = failure.Wrap("tsubaki.main.river.client.start", err)
		fmt.Fprintf(os.Stderr, "main: %v\n", err)
		errlog.Internal(riverRunCtx, nil, "river_client_start", failure.InternalGeneralError, err)
		return 1
	}
	slog.Info("River started")

	connCtx, connSpan := tracer.Start(rootCtx, "tsubaki.main.grpc.connection.start")
	defer connSpan.End()
	lis, err := net.Listen("tcp", gRPCAddr)
	if err != nil {
		err = failure.Wrap("tsubaki.main.listen.port.create", err)
		errlog.Internal(connCtx, nil, "connection_listen", failure.InternalGeneralError, err)
		return 1
	}

	handler := tsubakigrpc.New(queries, secretKey)
	grpcServer := grpc.NewServer()
	pb.RegisterRiskContextServiceServer(grpcServer, handler)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			errlog.Internal(rootCtx, nil, "grpc_serve", failure.InternalGeneralError, err)
		}
	}()

	<-ctx.Done()

	slog.Info("River/gRPC shutting down")
	defer func() {
		err := riverClient.Stop(ctx)
		if err != nil {
			err = failure.Wrap("main.river.client.stop", err)
			fmt.Fprintf(os.Stderr, "main: %v\n", err)
			errlog.Internal(riverRunCtx, nil, "river_client_stop", failure.InternalGeneralError, err)
		}
	}()

	return 0
}

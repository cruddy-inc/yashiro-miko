package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"pkg/errlog"
	"pkg/failure"
	"pkg/logger"
	db "tsubaki/internal/db/sqlc"

	"github.com/riverqueue/river"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
)

type CasesOpenedOutboxWorker struct {
	river.WorkerDefaults[CasesOpenedOutboxArgs]
	Queries *db.Queries
	Kafka   *kgo.Client
}

func (c *CasesOpenedOutboxWorker) Work(ctx context.Context, job *river.Job[CasesOpenedOutboxArgs]) error {
	log := slog.With(slog.String("op", "cases_opened_outbox"))
	ctx = logger.SetLogger(ctx, log)

	tracer := otel.Tracer("cases_opened_outbox")
	rootCtx, rootSpan := tracer.Start(ctx, "tsubaki.cases.opened.outbox.start")
	defer rootSpan.End()

	fetchKafkaCtx, fetchKafkaSpan := tracer.Start(rootCtx, "tsubaki.cases.opened.outbox.kafka.fetch")
	defer fetchKafkaSpan.End()

	fail, err := c.Queries.GetFailedCasesOpened(fetchKafkaCtx)
	if err != nil {
		err = failure.Wrap("tsubaki.cases.opened.outbox.failed.cases.opened.fetch", err)
		errlog.Database(fetchKafkaCtx, nil, "failed_cases_opened_fetch", err)
		return err
	}

	insertKafkaCtx, insertKafkaSpan := tracer.Start(rootCtx, "tsubaki.cases.opened.outbox.kafka.insert")
	defer insertKafkaSpan.End()
	for i := 0; i < len(fail); i++ {
		b := fail[i]

		if b.Status != "PENDING" {
			continue
		}

		record := &kgo.Record{
			Topic: b.KafkaTopic,
			Key:   b.KafkaCase,
			Value: b.Payload,
		}

		if err := c.Kafka.ProduceSync(insertKafkaCtx, record).FirstErr(); err != nil {
			log.Debug(fmt.Sprintf("Failed to insert Kafka record: %v", err))
			continue
		}

		c.Queries.InsertCasesOpenedPublished(insertKafkaCtx, b.CaseID)
	}

	return nil
}

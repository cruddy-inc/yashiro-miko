package httpy

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"pkg/errlog"
	"pkg/failure"
	"pkg/logger"
	"strconv"
	"strings"
	"time"
	"yaotose/internal/crypto"
	db "yaotose/internal/db/sqlc"
	"yaotose/internal/domain"
	yaotosev1 "yaotose/internal/gen/yaotose/v1"
	"yaotose/internal/normalize"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/proto"
)

func (y *Yaotose) IncomingTransaction(w http.ResponseWriter, r *http.Request) {
	var req domain.IncomingPayment
	pipe := y.Redis.TxPipeline()

	tracer := otel.Tracer("yaotose_http")
	ctx := context.Background()

	log := slog.With(slog.String("op", "yaotose_incoming_request"))
	ctx = logger.SetLogger(ctx, log)

	rootCtx, rootSpan := tracer.Start(ctx, "yaotose.http.incoming.transaction.start")
	defer rootSpan.End()

	tx, err := y.Pool.Begin(rootCtx)
	if err != nil {
		err = failure.Wrap("yaotose.http.pool.transaction.begin", err)
		errlog.General(rootCtx, w, "pool_transaction_begin", failure.InternalGeneralError, err)
		return
	}

	queries := y.Queries.WithTx(tx)
	defer tx.Rollback(rootCtx)

	err = sonic.ConfigDefault.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		err = failure.Wrap("yaotose.json.incoming.request.decode", err)
		errlog.Json(rootCtx, w, "request_decode", err)
		return
	}

	normalizeCtx, normalizeSpan := tracer.Start(rootCtx, "yaotose.http.transaction.data.normalize")
	defer normalizeSpan.End()
	payeeNormalized, err := normalize.NormalizeIBAN(req.PayeeIban)
	if err != nil {
		err = failure.Wrap("yaotose.http.payee.iban.normalize", err)
		errlog.General(normalizeCtx, w, "payee_iban_normalize", failure.InvalidError, err)
		return
	}

	payerNormalized, err := normalize.NormalizeIBAN(req.PayerIban)
	if err != nil {
		err = failure.Wrap("yaotose.http.payee.iban.normalize", err)
		errlog.General(normalizeCtx, w, "payee_iban_normalize", failure.InvalidError, err)
		return
	}

	secretKey := y.SecretKey.Load()
	if secretKey == nil {
		err := failure.InternalGeneral("secret_key_load", "empty secret key structure")
		err = failure.Wrap("yaotose.http.secret.key.load.check", err)
		errlog.Internal(normalizeCtx, w, "secret_key_load", failure.InternalEmptyValueError, err)
		return
	}

	payerToken := crypto.HMAC([]byte(secretKey.CurrentSecretKey), payerNormalized)

	amountCent := normalize.AmountToCent(req.Amount)

	c := normalize.AmountNew(y.Currency)

	amountEuro, err := c.AmountToEurCent(normalizeCtx, amountCent, normalize.Upper(req.Currency))
	if err != nil {
		err = failure.Wrap("yaotose.http.amount.convert.euro", err)
		errlog.Resource(normalizeCtx, w, "amount_convert_to_euro", failure.MalformedError, err)
		return
	}

	currPayeeToken := crypto.HMAC([]byte(secretKey.CurrentSecretKey), payeeNormalized)
	prevPayeeToken := crypto.HMAC([]byte(secretKey.PreviousSecretKey), payeeNormalized)

	blacklistCtx, blacklistSpan := tracer.Start(rootCtx, "yaotose.http.payee.blacklist.check")
	defer blacklistSpan.End()
	// Fail-Fast
	blacklistTimeout, cancelBlacklist := context.WithTimeout(blacklistCtx, 12*time.Millisecond)
	defer cancelBlacklist()
	_, err = pipe.Get(blacklistTimeout, "yaotose:payee:blacklist:"+currPayeeToken).Result()
	if err != nil {
		_, err = pipe.Get(blacklistTimeout, "yaotose:payee:blacklist:"+prevPayeeToken).Result()
	}
	switch {
	// payee is not blacklisted
	case err == redis.Nil:
		// Continue workflow
	// payee is blacklisted
	case err == nil:
		err := failure.Forbidden("transaction", "payee is blacklisted")
		err = failure.Wrap("yaotose.redis.payee.blacklist.check", err)
		errlog.Resource(blacklistCtx, w, "payee_blacklist_check", failure.ForbiddenError, err)
		return
	// Redis might be down, fallback to postgres
	case err != nil:
		postgresTimeout, cancelPSLQ := context.WithTimeout(blacklistCtx, 25*time.Millisecond)
		defer cancelPSLQ()
		exists, err := queries.GetBlacklist(postgresTimeout, db.GetBlacklistParams{
			CurrIbanHash: currPayeeToken,
			PrevIbanHash: prevPayeeToken,
			TenantID:     req.TenantID,
		})
		if err != nil {
			err = failure.Wrap("yaotose.database.payee.blacklist.check", err)
			errlog.Database(blacklistCtx, w, "payee_blacklist_check", err)
			return
		}
		if exists {
			err := failure.Forbidden("transaction", "payee is blacklisted")
			err = failure.Wrap("yaotose.redis.payee.blacklist.check", err)
			errlog.Resource(blacklistCtx, w, "payee_blacklist_check", failure.ForbiddenError, err)
			return
		}
		// Nothing matches, payee is not blacklisted
	}

	key := fmt.Sprintf("yaotose:%s:%s:%s:last20", string(req.TenantID), payerToken, currPayeeToken)

	historyCtx, historySpan := tracer.Start(rootCtx, "yaotose.http.transaction.history.fetch")
	defer historySpan.End()
	rulesTimeout, cancelRules := context.WithTimeout(historyCtx, 10*time.Millisecond)
	defer cancelRules()

	// We get a slice of the values
	data := pipe.LRange(rulesTimeout, key, 0, 19)
	if err != nil {
		err = failure.Wrap("yaotose.redis.transaction.history24.fetch", err)
		errlog.Redis(historyCtx, w, "payer_payee_history24_fetch", err)
		return
	}

	var transactionsWithin24 int
	var amountWithin24 int64
	var transactionsWithin1 int
	var amountWithin1 int64

	items, err := data.Result()
	if err != nil {
		err = failure.InternalGeneral("redis_data_parse", "failed to parse redis data")
		errlog.Redis(historyCtx, w, "redis_data_parse", err)
		return
	}

	oneHourThreshold := time.Now().Add(time.Hour).Unix()

	for _, item := range items {
		t, a, found := strings.Cut(item, ":")
		if !found {
			continue
		}
		ti, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			continue
		}
		am, err := strconv.ParseInt(a, 10, 64)
		if err != nil {
			continue
		}

		if ti <= oneHourThreshold {
			transactionsWithin1++
			amountWithin1 += am
		}

		transactionsWithin24++
		amountWithin24 += am
	}

	knownPayee := true
	info, err := queries.GetPayerPayee(historyCtx, db.GetPayerPayeeParams{
		TenantID:  req.TenantID,
		PayerHash: payerToken,
		PayeeHash: currPayeeToken,
	})
	if err != nil {
		knownPayee = false
	}

	evalCtx, evalSpan := tracer.Start(rootCtx, "yaotose.http.cel.decision.evaluate")
	defer evalSpan.End()
	verdict, err := y.Engine.Evaluate(map[string]any{
		"remote_access_detected": req.RemoteAccessDetected,
		"active_phone_call":      req.ActivePhoneCall,
		"known_payee":            knownPayee,
		"amount_eur_minor":       amountEuro,
		"payee_verification":     req.PayeeVerification,
		"n_payments":             info.NPayments,
		"max_minor":              info.MaxAmount,
		"now":                    time.Now().Unix(),
		"first_transaction":      info.FirstTransaction,
		"last_transaction":       info.LastTransaction,
		"segment":                req.Segment,
		"count_1h":               transactionsWithin1,
		"sum_1h":                 amountWithin1,
		"count_24h":              transactionsWithin24,
		"sum_24h":                amountWithin24,
	})
	if err != nil {
		err = failure.Wrap("yaotose.cel.when.condition.evaluate", err)
		errlog.General(evalCtx, w, "cel_rules_evaluate", failure.InternalGeneralError, err)
		return
	}

	if verdict.Decision != "ALLOW" {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		s := sonic.ConfigStd.NewEncoder(w)
		_ = s.Encode(map[string]string{
			"message": "Transaction has been stopped due to suspiciou activities, please, contact the operator if it's a misunderstanding",
		})
	}

	outboxCtx, outboxSpan := tracer.Start(rootCtx, "yaotose.http.cases.opened.publish")
	defer outboxSpan.End()

	caseId := uuid.Must(uuid.NewV7()).String()
	caseForm := fmt.Sprintf("case.%s.%s.%s", req.TenantID, payerToken, currPayeeToken)
	caseByte := []byte(caseForm)

	var payerToPaye *yaotosev1.PayerToPayeeInteraction
	if knownPayee {
		payerToPaye = &yaotosev1.PayerToPayeeInteraction{
			KnownPayee:                   knownPayee,
			NPayments:                    info.NPayments,
			MaxAmountSentMinor:           int32(info.MaxAmount),
			TransactionsWithin_1Hour:     int32(transactionsWithin1),
			AmountSentWithin_1HourMinor:  int32(amountWithin1),
			TransactionsWithin_24Hour:    int32(transactionsWithin24),
			AmountSentWithin_24HourMinor: int32(amountWithin24),
			FirstTransactionUnix:         info.FirstTransaction,
			LastTransactionUnix:          info.LastTransaction,
		}
	}

	value := yaotosev1.Envelope{
		CaseId:          caseId,
		TenantId:        req.TenantID,
		PayerIbanHash:   payerToken,
		PayeeIbanHash:   currPayeeToken,
		YaotoseDecision: verdict.Decision,
		RulesFired: &yaotosev1.YaotoseData{
			Rules:  verdict.SuspicionAlert,
			Weight: int32(verdict.Weight),
		},
		CurrentTransactionInfo: &yaotosev1.CurrentTransactionInfo{
			Segment:              req.Segment,
			UntrustedMemo:        req.RemittanceInfo,
			OriginalCurrency:     req.Currency,
			AmountEurMinor:       amountEuro,
			ActivePhoneCall:      req.ActivePhoneCall,
			RemoteAccessDetected: req.RemoteAccessDetected,
		},
		PayerToPayeeInteraction: payerToPaye,
		KafkaTopic:              "cases.opened",
		KafkaCase:               caseByte,
	}

	payload, err := proto.Marshal(&value)
	if err != nil {
		err = failure.Wrap("yaotose.proto.yaotosev1.envelope.marshal", err)
		errlog.General(outboxCtx, nil, "yaotosev1_envelope_proto_marshal", failure.InternalGeneralError, err)
		return
	}

	rows, err := queries.InsertCasesOpened(outboxCtx, db.InsertCasesOpenedParams{
		CaseID:     caseId,
		KafkaTopic: "cases.opened",
		KafkaCase:  caseByte,
		Payload:    payload,
		Status:     "PENDING",
	})
	if err != nil {
		err = failure.Wrap("yaotose.database.cases.opened.insert", err)
		errlog.Database(outboxCtx, nil, "cases_opened_insert", err)
		return
	}
	if rows == 0 {
		err = failure.Wrap("yaotose.database.cases.opened.rate.limit.check", failure.RateLimit())
		errlog.Resource(outboxCtx, nil, "cases_opened_insert", failure.RateLimitError, err)
		return
	}

	tx.Commit(outboxCtx)

	record := &kgo.Record{
		Topic: "cases.opened",
		Key:   []byte(caseForm),
		Value: payload,
	}

	// TODO: Don't forget to create a background process to monitor what didn't get sent to kafka
	Just_A_Planned_Error

	kafkaTimeout, kafkaCancel := context.WithTimeout(outboxCtx, 2*time.Second)
	defer kafkaCancel()

	if err := y.Kafka.ProduceSync(kafkaTimeout, record).FirstErr(); err != nil {
		err = failure.Wrap("yaotose.kafka.cases.opened.produce.sync", err)
		errlog.General(outboxCtx, nil, "kafka_produce_sync", failure.InternalGeneralError, err)
		return
	} else {
		_ = queries.InsertCasesOpenedPublished(outboxCtx, caseId)
	}
}

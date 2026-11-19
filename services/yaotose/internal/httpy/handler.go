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
	"yaotose/internal/normalize"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
)

func (y *Yaotose) IncomingTransaction(w http.ResponseWriter, r *http.Request) {
	var req domain.IncomingPayment
	pipe := y.Redis.TxPipeline()

	tracer := otel.Tracer("incoming_transaction")
	ctx := context.Background()

	log := slog.With(slog.String("op", "yaotose_incoming_request"))
	ctx = logger.SetLogger(ctx, log)

	rootCtx, rootSpan := tracer.Start(ctx, "incoming.transaction.start")
	defer rootSpan.End()

	err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		err = failure.Wrap("yaotose.json.incoming.request.decode", err)
		errlog.Json(rootCtx, w, "request_decode", err)
		return
	}

	polishCtx, polishSpan := tracer.Start(rootCtx, "incoming.transaction.polish.iban")
	defer polishSpan.End()
	payeeNormalized, err := normalize.NormalizeIBAN(req.PayeeIban)
	if err != nil {
		err = failure.Wrap("yaotose.http.payee.iban.normalize", err)
		errlog.General(polishCtx, w, "payee_iban_normalize", failure.InvalidError, err)
		return
	}

	payerNormalized, err := normalize.NormalizeIBAN(req.PayerIban)
	if err != nil {
		err = failure.Wrap("yaotose.http.payee.iban.normalize", err)
		errlog.General(polishCtx, w, "payee_iban_normalize", failure.InvalidError, err)
		return
	}

	payerToken := crypto.HMAC([]byte(y.SecretKey), payerNormalized)

	amountCent := normalize.AmountToCent(req.Amount)

	amountEuro, err := normalize.AmountToEurCent(polishCtx, amountCent, normalize.Upper(req.Currency))
	if err != nil {
		err = failure.Wrap("yaotose.http.amount.convert.euro", err)
		errlog.Resource(polishCtx, w, "amount_convert_to_euro", failure.MalformedError, err)
		return
	}

	payeeToken := crypto.HMAC([]byte(y.SecretKey), payeeNormalized)

	checkCtx, checkSpan := tracer.Start(polishCtx, "incoming.transaction.payee.check")
	defer checkSpan.End()
	// Fail-Fast
	blacklistTimeout, cancelBlacklist := context.WithTimeout(checkCtx, 10*time.Millisecond)
	defer cancelBlacklist()
	_, err = pipe.Get(blacklistTimeout, "yaotose:payee:blacklist:"+string(payeeToken)).Result()
	switch {
	// payee is not blacklisted
	case err == redis.Nil:
		// Continue workflow
	// payee is blacklisted
	case err == nil:
		err := failure.Forbidden("transaction", "payee is blacklisted")
		err = failure.Wrap("yaotose.redis.payee.blacklist.check", err)
		errlog.Resource(checkCtx, w, "payee_blacklist_check", failure.ForbiddenError, err)
		return
	// Redis might be down, fallback to postgres
	case err != nil:
		postgresTimeout, cancelPSLQ := context.WithTimeout(checkCtx, 25*time.Millisecond)
		defer cancelPSLQ()
		exists, err := y.Queries.GetBlacklist(postgresTimeout, db.GetBlacklistParams{
			IbanHash: string(payeeToken),
			TenantID: req.TenantID,
		})
		if err != nil {
			err = failure.Wrap("yaotose.database.payee.blacklist.check", err)
			errlog.Database(checkCtx, w, "payee_blacklist_check", err)
			return
		}
		if exists {
			err := failure.Forbidden("transaction", "payee is blacklisted")
			err = failure.Wrap("yaotose.redis.payee.blacklist.check", err)
			errlog.Resource(checkCtx, w, "payee_blacklist_check", failure.ForbiddenError, err)
			return
		}
		// Nothing matches, payee is not blacklisted
	}

	key := fmt.Sprintf("yaotose:%s:%s:%s:last20", string(req.TenantID), string(payerToken), string(payeeToken))

	transactionCtx, transactionSpan := tracer.Start(rootCtx, "incoming.transaction.info.fetch")
	defer transactionSpan.End()
	rulesTimeout, cancelRules := context.WithTimeout(transactionCtx, 10*time.Millisecond)
	defer cancelRules()
	// We get a slice of the values
	data := pipe.LRange(rulesTimeout, key, 0, 19)
	if err != nil {
		err = failure.Wrap("yaotose.redis.transaction.history24.fetch", err)
		errlog.Redis(transactionCtx, w, "payer_payee_history24_fetch", err)
		return
	}

	var transactionsWithin24 int
	var amountWithin24 int64
	var transactionsWithin1 int
	var amountWithin1 int64

	items, err := data.Result()
	if err != nil {
		err = failure.InternalGeneral("redis_data_parse", "failed to parse redis data")
		errlog.Redis(transactionCtx, w, "redis_data_parse", err)
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
	info, err := y.Queries.GetPayerPayee(transactionCtx, db.GetPayerPayeeParams{
		TenantID:  req.TenantID,
		PayerHash: string(payerToken),
		PayeeHash: string(payeeToken),
	})
	if err != nil {
		knownPayee = false
	}

	finaleCtx, finaleSpan := tracer.Start(rootCtx, "incoming.transaction.evaluate.decision")
	defer finaleSpan.End()
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
		errlog.General(finaleCtx, w, "cel_rules_evaluate", failure.InternalGeneralError, err)
		return
	}

	if verdict.Decision != "ALLOW" {
		// Kafka
	}

	// Else we continue
}

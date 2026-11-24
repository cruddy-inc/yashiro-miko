package currency

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"pkg/errlog"
	"pkg/failure"
	"pkg/logger"
	"time"
	"tsubaki/domain"

	"github.com/bytedance/sonic"
	"github.com/riverqueue/river"
	"go.opentelemetry.io/otel"
)

type CurrencyTrackWorker struct {
	river.WorkerDefaults[CurrencyTrackArgs]
	HttpClient domain.HttpDo
	Rates      *domain.CurrencyStore
}

type Currency string

const currency Currency = "AED,AUD,BDT,BGN,BWP,CAD,CHF,CLP,CNY,CRC,CZK,DKK,EGP,EUR,GBP,GEL,GHS,HKD,HUF,IDR,ILS,INR,JPY,KES,KRW,LKR,MAD,MXN,MYR,NGN,NOK,NPR,NZD,PHP,PKR,PLN,RON,SEK,SGD,THB,TRY,TZS,UAH,UGX,USD,UYU,VND,XOF,ZAR,ZMW"

func (c *CurrencyTrackWorker) Work(ctx context.Context, job *river.Job[CurrencyTrackArgs]) error {
	tracer := otel.Tracer("currency_tarcker")

	log := slog.With(slog.String("op", "currency_tracker"))
	ctx = logger.SetLogger(ctx, log)

	rootCtx, rootSpan := tracer.Start(ctx, "tsubaki.currency.tracker.start")
	defer rootSpan.End()

	requestCtx, requestSpan := tracer.Start(rootCtx, "tsubaki.currency.tracker.request")
	defer requestSpan.End()
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", "https://api.frankfurter.dev/v2/rates?base=EUR&quotes=", currency), nil)
	if err != nil {
		err = failure.Wrap("tsubaki.currency.tracker.request.create", err)
		errlog.General(requestCtx, nil, "request_create", failure.InternalGeneralError, err)
		return err
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		err = failure.Wrap("tsubaki.currency.tracker.request.send", err)
		errlog.General(requestCtx, nil, "request_send", failure.InternalGeneralError, err)
		return err
	}

	responseCtx, responseSpan := tracer.Start(rootCtx, "tsubaki.currency.tracker.response")
	defer responseSpan.End()
	var raw []domain.FrankfurterRate
	if err := sonic.ConfigDefault.NewDecoder(resp.Body).Decode(&raw); err != nil {
		err = failure.Wrap("tsubaki.currency.tracker.response.body.decode", err)
		errlog.Json(responseCtx, nil, "response_body_decode", err)
		return err
	}

	insertCtx, insertSpan := tracer.Start(rootCtx, "tsubaki.currency.tracker.parse")
	defer insertSpan.End()
	var t string
	if len(raw) > 0 {
		t = raw[0].Date
	}

	rates := make(map[string]float64, len(raw)+1)
	for _, item := range raw {
		rates[item.Quote] = item.Rate
	}

	ti, err := time.Parse("2006-01-02", t)
	if err != nil {
		err = failure.Wrap("tsubaki.currency.tracker.time.parse", err)
		errlog.Resource(insertCtx, nil, "time_parse", failure.MalformedError, err)
		return err
	}

	c.Rates.Store(
		&domain.Currency{
			Rate:      rates,
			UpdatedAt: ti,
		},
	)

	return nil
}

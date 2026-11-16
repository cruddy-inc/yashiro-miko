package errlog

import (
	"context"
	"errors"
	"net/http"
	"pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func Redis(ctx context.Context, subject string, err error) {

	logger := logger.GetLogger(ctx)
	span := trace.SpanFromContext(ctx)
	red, isRedisError := errors.AsType[redis.Error](err)

	var code string
	var failed bool
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		failed = true
		code = "redis_timeout"
		logger.Error("context deadline exceeded",
			"err", err,
			"status", http.StatusGatewayTimeout,
			"code", code,
		)
	case errors.Is(err, context.Canceled):
		logger.Warn("redis operation canceled",
			"err", err,
			"status", http.StatusInternalServerError,
			"code", "redis_operation_canceled",
		)
	case redis.IsAuthError(err):
		failed = true
		code = "redis_authentication_failed"
		logger.Error("failed redis authentication",
			"err", err,
			"status", http.StatusInternalServerError,
			"code", code,
		)
	case errors.Is(err, redis.Nil):
		logger.Warn("expired or nonexistent redis entry",
			"err", err,
			"status", http.StatusNotFound,
			"code", subject+"_not_found",
		)
	case isRedisError:
		failed = true
		code = subject + "_failed"
		logger.Error("redis server rejected command: "+red.Error(),
			"err", err,
			"status", http.StatusInternalServerError,
			"code", code,
		)
	default:
		failed = true
		code = subject + "_failed"
		logger.Error("unexpected redis error",
			"err", err,
			"status", http.StatusInternalServerError,
			"code", code,
		)
	}

	if failed {
		span.RecordError(err)
		span.SetStatus(codes.Error, code)
	}
}

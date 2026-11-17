package errlog

import (
	"context"
	"net/http"
	"pkg/failure"
	"pkg/logger"
	"strings"
)

func Resource(ctx context.Context, w http.ResponseWriter, subject string, kind failure.KindType, err error) {
	logger := logger.GetLogger(ctx)
	subjectMsg := strings.ReplaceAll(subject, "_", " ")

	switch kind {
	case failure.EmptyValueError:
		logger.WarnContext(ctx, subjectMsg+" is missing",
			"err", err,
			"code", subject+"_missing",
		)
		failure.Response(w, failure.EmptyValueError)
	case failure.MalformedError:
		logger.WarnContext(ctx, "resource is malformed",
			"err", err,
			"code", subject+"_malformed",
		)
		failure.Response(w, failure.MalformedError)
	case failure.InvalidError:
		logger.WarnContext(ctx, "request validation failed",
			"err", err,
			"code", subject+"_invalid",
		)
		failure.Response(w, failure.InvalidError)
	case failure.UnauthorizedError:
		logger.WarnContext(ctx, "unauthorized "+subjectMsg,
			"err", err,
			"code", subject+"_unauthorized",
		)
		failure.Response(w, failure.UnauthorizedError)
	case failure.ForbiddenError:
		logger.WarnContext(ctx, "not enough permissions for "+subjectMsg,
			"err", err,
			"code", subject+"_forbidden",
		)
		failure.Response(w, failure.ForbiddenError)
	case failure.ConflictError:
		logger.WarnContext(ctx, "resource conflict",
			"err", err,
			"code", subject+"_already_exists",
		)
		failure.Response(w, failure.ConflictError)
	case failure.NotFoundError:
		logger.WarnContext(ctx, subjectMsg+" not found",
			"err", err,
			"code", subject+"_not_found",
		)
		failure.Response(w, failure.NotFoundError)
	case failure.RateLimitError:
		logger.WarnContext(ctx, subjectMsg+" is rate limited",
			"err", err,
			"code", subject+"_rate_limited",
		)
		failure.Response(w, failure.RateLimitError)
	}
}

package failure

import (
	"net/http"

	"github.com/bytedance/sonic"
)

type Resp struct {
	Code    string
	Message string
}

func Response(w http.ResponseWriter, kind KindType) {
	if w == nil {
		return
	}
	var resp *Resp
	var status int

	w.Header().Set("Content-Type", "application/json")
	switch kind {
	case EmptyValueError:
		resp = &Resp{
			Code:    "EMPTY_VALUE",
			Message: "Please fill in all required fields and try again.",
		}
		status = http.StatusBadRequest
	case MalformedError:
		resp = &Resp{
			Code:    "MALFORMED",
			Message: "The request format is invalid. Please try again",
		}
		status = http.StatusBadRequest
	case InvalidError:
		resp = &Resp{
			Code:    "INVALID",
			Message: "One or more entered details are invalid. Please review and try again.",
		}
		status = http.StatusBadRequest
	case UnauthorizedError:
		resp = &Resp{
			Code:    "UNAUTHORIZED",
			Message: "Not enough data to perform action. Please sign in.",
		}
		status = http.StatusUnauthorized
	case RateLimitError:
		resp = &Resp{
			Code:    "RATE_LIMITED",
			Message: "Too many attempts. Please wait a moment before trying again.",
		}
		status = http.StatusTooManyRequests
	case ConflictError:
		resp = &Resp{
			Code:    "CONFLICT",
			Message: "This transaction is already being processed or has already been completed.",
		}
		status = http.StatusConflict
	case ForbiddenError, NotFoundError, GeneralError:
		resp = &Resp{
			Code:    "TRANSACTION_DECLINED",
			Message: "Transaction declined. For specific inquiries, please contact our support service.",
		}
		status = http.StatusUnprocessableEntity
	default:
		resp = &Resp{
			Code:    "TRANSACTION_FAILED",
			Message: "Transaction failed. For specific inquiries, please contact our support service.",
		}
		status = http.StatusInternalServerError
	}

	w.WriteHeader(status)
	s := sonic.ConfigDefault.NewEncoder(w)
	s.SetIndent("", "    ")
	_ = s.Encode(resp)
}

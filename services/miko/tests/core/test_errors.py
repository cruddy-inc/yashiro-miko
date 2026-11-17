from miko.core.logging.kinds import ErrorKind
from miko.core.logging.errors import (
    AppError,
    InvalidError,
    NotFoundError,
    UnauthorizedError,
    ForbiddenError,
    ConflictError,
    RateLimitError,
    InternalError
)


def test_app_error_defaults():
    err = AppError()
    assert err.message == "An internal error occurred."
    assert err.extra == {}


def test_app_custom_message_and_extras():
    err = AppError(
        "This is a custom message by Shrimp", node_id="node_12", retries=3, interval=2
    )
    assert err.message == "This is a custom message by Shrimp"
    assert err.extra == {"node_id": "node_12", "retries": 3, "interval": 2}


def test_subclass_override_defaults():
    err = NotFoundError(resource_id="res_trial")
    assert err.status_code == 404
    assert err.message == "The request resource was not found"
    assert err.extra == {"resource_id": "res_trial"}


def test_kind_of_error():
    err = InvalidError()
    assert err.kind == ErrorKind.INVALID


def test_cause_chaining():
    cause = ValueError("Invalid UUID")
    err = AppError("Parsing failed", cause=cause)
    assert err.__cause__ is cause

def test_error_subclasses():
    cases = [
        (InvalidError, ErrorKind.INVALID, 400),
        (UnauthorizedError, ErrorKind.UNAUTHORIZED, 401),
        (ForbiddenError, ErrorKind.FORBIDDEN, 403),
        (NotFoundError, ErrorKind.NOT_FOUND, 404),
        (ConflictError, ErrorKind.CONFLICT, 409),
        (RateLimitError, ErrorKind.RATE_LIMIT, 429),
        (InternalError, ErrorKind.INTERNAL, 500),
    ]

    for error_class, expected_kind, expected_status in cases:
        err = error_class()

        assert err.kind == expected_kind
        assert err.status_code == expected_status
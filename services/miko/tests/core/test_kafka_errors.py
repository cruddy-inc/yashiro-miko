from miko.core.logging.handlers import (
    KafkaFailureDisposition,
    from_kafka_exceptions,
    from_kafka_error,
    classify_kafka_errors,
)

from unittest.mock import MagicMock
from confluent_kafka import KafkaError, KafkaException
import pytest


@pytest.mark.parametrize(
    ("is_retriable", "is_fatal", "is_txn_abort", "expected"),
    [
        (True, False, False, KafkaFailureDisposition.RETRYABLE),
        (False, True, False, KafkaFailureDisposition.FATAL),
        (False, False, True, KafkaFailureDisposition.TRANSACTION_ABORT_REQUIRED),
        (False, False, False, KafkaFailureDisposition.NON_RETRYABLE),
    ],
)
def test_classify_kafka_errors(
    is_retriable: bool,
    is_fatal: bool,
    is_txn_abort: bool,
    expected: KafkaFailureDisposition,
) -> None:
    mock_err = MagicMock(spec=KafkaError)
    mock_err.retriable.return_value = is_retriable
    mock_err.fatal.return_value = is_fatal
    mock_err.txn_requires_abort.return_value = is_txn_abort

    assert classify_kafka_errors(mock_err) == expected


def test_classify_kafka_error_none_returns_unknown() -> None:
    assert classify_kafka_errors(None) == KafkaFailureDisposition.UNKNOWN


@pytest.mark.parametrize(
    ("operation", "topic", "partition", "offset"),
    [
        ("miko.kafka.seek", "transaction-events", 101, 1),
        ("miko.kafka.consume", "risk-assessment", 2, 404),
        ("miko.kafka.produce", "audit-ledger", 0, 12),
        ("miko.kafka.commit", None, None, None),
    ],
)
def test_from_kafka_error(
    operation: str,
    topic: str | None,
    partition: int | None,
    offset: int | None,
) -> None:
    mock_err = MagicMock(spec=KafkaError)
    mock_err.retriable.return_value = False
    mock_err.fatal.return_value = False
    mock_err.txn_requires_abort.return_value = False

    err = from_kafka_error(
        mock_err,
        operation=operation,
        topic=topic,
        partition=partition,
        offset=offset,
    )

    assert err.operation == operation
    assert err.extra.get("topic") == topic
    assert err.extra.get("partition") == partition
    assert err.extra.get("offset") == offset


@pytest.mark.parametrize(
    ("has_inner_kafka_error", "operation", "topic", "partition", "offset"),
    [
        (True, "miko.kafka.poll", "transaction-events", 3, 512),
        (False, "miko.kafka.flush", "audit-ledger", 0, 12),
        (True, "miko.kafka.seek", "risk-assessment", 101, 1),
        (False, "miko.kafka.commit", None, None, None),
    ],
)
def test_from_kafka_exceptions(
    has_inner_kafka_error: bool,
    operation: str,
    topic: str | None,
    partition: int | None,
    offset: int | None,
) -> None:
    if has_inner_kafka_error:
        mock_kafka_err = MagicMock(spec=KafkaError)
        mock_kafka_err.retriable.return_value = False
        mock_kafka_err.fatal.return_value = False
        mock_kafka_err.txn_requires_abort.return_value = False
        exc = KafkaException(mock_kafka_err)
    else:
        mock_kafka_err = None
        exc = KafkaException("Generic failure message")

    err = from_kafka_exceptions(
        exc,
        operation=operation,
        topic=topic,
        partition=partition,
        offset=offset,
    )

    assert err.kafka_error is mock_kafka_err
    assert err.__cause__ is exc
    assert err.operation == operation
    assert err.extra.get("topic") == topic
    assert err.extra.get("partition") == partition
    assert err.extra.get("offset") == offset

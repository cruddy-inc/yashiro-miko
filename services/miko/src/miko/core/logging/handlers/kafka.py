from enum import StrEnum

from confluent_kafka import KafkaError, KafkaException

from miko.core.logging.errors import AppError
from miko.core.logging.kinds import ErrorKind


class KafkaFailureDisposition(StrEnum):
    UNKNOWN = "unknown"
    RETRYABLE = "retryable"
    NON_RETRYABLE = "non_retryable"
    FATAL = "fatal"
    TRANSACTION_ABORT_REQUIRED = "transaction_abort_required"


def classify_kafka_errors(
    error: KafkaError | None,
) -> KafkaFailureDisposition:

    if error is None:
        return KafkaFailureDisposition.UNKNOWN

    if error.retriable():
        return KafkaFailureDisposition.RETRYABLE

    if error.fatal():
        return KafkaFailureDisposition.FATAL

    if error.txn_requires_abort():
        return KafkaFailureDisposition.TRANSACTION_ABORT_REQUIRED

    return KafkaFailureDisposition.NON_RETRYABLE


class KafkaOperationError(AppError):
    kind = ErrorKind.INTERNAL
    default_message = "A Kafka operation failed"

    def __init__(
        self,
        message: str | None = None,
        *,
        operation: str,
        kafka_error: KafkaError | None = None,
        topic: str | None = None,
        partition: int | None = None,
        offset: int | None = None,
        cause: Exception | None = None,
    ) -> None:

        self.operation = operation
        self.kafka_error = kafka_error
        self.disposition = classify_kafka_errors(kafka_error)

        metadata: dict[str, object] = {
            "operation": operation,
            "disposition": self.disposition,
        }

        if kafka_error is not None:
            metadata["kafka_code"] = kafka_error.code()
            metadata["kafka_name"] = kafka_error.name()

        if topic is not None:
            metadata["topic"] = topic

        if partition is not None:
            metadata["partition"] = partition

        if offset is not None:
            metadata["offset"] = offset

        super().__init__(message, cause=cause, **metadata)


def from_kafka_error(
    error: KafkaError,
    *,
    operation: str,
    topic: str | None = None,
    partition: int | None = None,
    offset: int | None = None,
) -> KafkaOperationError:

    return KafkaOperationError(
        message=f"Kafka operation failed: {operation}",
        operation=operation,
        kafka_error=error,
        topic=topic,
        partition=partition,
        offset=offset,
    )


def from_kafka_exceptions(
    exc: KafkaException,
    *,
    operation: str,
    topic: str | None = None,
    partition: int | None = None,
    offset: int | None = None,
) -> KafkaOperationError:
    kafka_error = None

    if exc.args and isinstance(exc.args[0], KafkaError):
        kafka_error = exc.args[0]

    return KafkaOperationError(
        message=f"Kafka operation failed: {operation}",
        operation=operation,
        kafka_error=kafka_error,
        topic=topic,
        partition=partition,
        offset=offset,
        cause=exc,
    )

from miko.core.logging.errors import AppError, InternalError

def wrap(operation: str, err: Exception | None) -> AppError | None:

    if err is None:
        return None

    if isinstance(err, AppError):
        new_message = f"{operation}: {err.message}"

        wrapped_err = err.__class__(
            message=new_message,
            cause=err,
            **err.extra,
        )

        return wrapped_err

    return InternalError(
        message=f"{operation}: {str(err)}",
        cause=err,
    )

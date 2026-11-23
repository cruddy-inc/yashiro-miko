from .kinds import ErrorKind


# ==============================================================================
# ROOT APPLICATION EXCEPTION
# ==============================================================================

class AppError(Exception):
    kind: ErrorKind
    status_code: int = 500
    default_message: str = "An internal error occurred."

    def __init__(
        self,
        message: str | None = None,
        *,
        cause: Exception | None = None,
        **extra: object
    ) -> None:

        self.message = message if message is not None else self.default_message
        self.extra = extra

        super().__init__(self.message)

        if cause is not None:
            self.__cause__ = cause



# ==============================================================================
# HTTP / API TRANSPORT ERRORS
# ==============================================================================

class InvalidError(AppError):
    kind = ErrorKind.INVALID
    status_code = 400
    default_message = "Invalid request payload or arguments"


class UnauthorizedError(AppError):
    kind = ErrorKind.UNAUTHORIZED
    status_code = 401
    default_message = "Authentication is required"


class ForbiddenError(AppError):
    kind = ErrorKind.FORBIDDEN
    status_code = 403
    default_message = "You do not have the permissions to perform this action"


class NotFoundError(AppError):
    kind = ErrorKind.NOT_FOUND
    status_code = 404
    default_message = "The request resource was not found"


class ConflictError(AppError):
    kind = ErrorKind.CONFLICT
    status_code = 409
    default_message = "A resource conflict occurred"


class RateLimitError(AppError):
    kind = ErrorKind.RATE_LIMIT
    status_code = 429
    default_message = "Rate limit exceeded. Please try again later"


class InternalError(AppError):
    kind = ErrorKind.INTERNAL
    status_code = 500
    default_message = "An unexpected server error occurred"

# ==============================================================================
# DOMAIN / FEATURE SPECIFIC ERRORS (EXAMPLES)
# ==============================================================================

class DirectoryNotFoundError(NotFoundError):
    default_message = "The requested directory does not exist."


class PathIsNotADirectoryError(InvalidError):
    default_message = "The specified path is not a directory."


class InsufficientPermissionsError(ForbiddenError):
    default_message = "You do not have sufficient permissions to execute this action."


class ResourceFileNotFoundError(NotFoundError):
    default_message = "The requested file does not exist."
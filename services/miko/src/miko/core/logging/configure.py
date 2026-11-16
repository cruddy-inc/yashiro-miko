import logging
import sys

ROOT_LOGGER_NAME = "miko"
LOG_LEVEL = logging.INFO

LOG_FORMAT = (
    "%(asctime)s | %(levelname)-8s | %(name)-12s : %(message)s"
)


def configure_logging() -> None:
    logger = logging.getLogger(ROOT_LOGGER_NAME)

    if not logger.handlers:
        logger.setLevel(LOG_LEVEL)

        handler = logging.StreamHandler(sys.stdout)
        handler.setLevel(LOG_LEVEL)

        formatter = logging.Formatter(
            LOG_FORMAT,
            datefmt="%Y-%m-%d %H:%M:%S",
        )

        handler.setFormatter(formatter)
        logger.addHandler(handler)

    logger.propagate = False
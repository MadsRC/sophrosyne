"""Logging provides utilities that allow the app to standardise on logging."""

import logging
import random
import sys
import time
from typing import Any, Literal

import structlog
from fastapi import Request, Response
from starlette.middleware.base import BaseHTTPMiddleware


def get_logger() -> Any:
    """Get a logger.

    Primary purpose of this function is to not have other modules import
    structlog.
    """
    return structlog.get_logger()


def initialize_logging(
    log_level: int = logging.NOTSET,
    clear_handlers: bool = True,
    event_field: str = "event",
    format: Literal["development", "production"] = "production",
) -> None:
    """Set up logging.

    Will set up logging using structlog. In addition to setting up structlog,
    the root logger of the Python standard library (`logging`) will be
    reconfigured to use the same formatting as structlog.

    All logs will be written to standard out as JSON.

    If the `log_format` attribute of the `sophrosyne.core.config.Settings` class
    equals `development`, then logs will be pretty printed.

    By default, all handlers of the Python standard library logging package
    will be cleared, unless the `clear_handlers` argument is set to `False`.

    Additionally, the function will set clear the handlers and set logger to
    propagate for the following modules:

    - `uvicorn`
    - `uvicorn.error`

    This is in order to prevent them from setting their own format.

    Args:
        log_level (int): Logging level to use. Defaults to logging.NOTSET.
        clear_handlers (bool): Existing handlers should be cleared. Defaults to True.
        event_field (str): The name of the field that will contain the logged event. Defaults to "event".
        format (Literal["development", "production"]): Decides how the log will be formatted.
    """
    if clear_handlers:
        logging.getLogger().handlers.clear()

    processors: list[structlog.types.Processor] = [
        structlog.contextvars.merge_contextvars,
        structlog.stdlib.add_logger_name,
        structlog.stdlib.add_log_level,
        structlog.processors.StackInfoRenderer(),
        structlog.processors.TimeStamper(fmt="iso"),
    ]

    if event_field != "event":
        processors.append(structlog.processors.EventRenamer(to=event_field))

    log_renderer: structlog.typing.Processor
    if format == "development":
        processors.append(structlog.dev.set_exc_info)
        log_renderer = structlog.dev.ConsoleRenderer(event_key=event_field)
    else:
        processors.append(structlog.processors.format_exc_info)
        log_renderer = structlog.processors.JSONRenderer()

    structlog.configure(
        processors=processors
        + [structlog.stdlib.ProcessorFormatter.wrap_for_formatter],
        cache_logger_on_first_use=True,
        logger_factory=structlog.stdlib.LoggerFactory(),
    )

    formatter = structlog.stdlib.ProcessorFormatter(
        processors=[
            structlog.stdlib.ProcessorFormatter.remove_processors_meta,
            log_renderer,
        ],
        foreign_pre_chain=processors,
    )

    h = logging.StreamHandler()
    h.setFormatter(formatter)

    r = logging.getLogger()
    r.addHandler(h)
    r.setLevel(log_level)

    for name in ["uvicorn", "uvicorn.error"]:
        logging.getLogger(name).handlers.clear()
        logging.getLogger(name).propagate = True

    # Silence the uvicorn access logger, as we reimplement this ourselves.
    logging.getLogger("uvicorn.access").handlers.clear()
    logging.getLogger("uvicorn.access").propagate = False

    def handle_exception(exc_type, exc_value, exc_traceback):
        """Log any uncaught exception.

        Ignores KeyboardInterrupt from Ctrl+C.
        """
        if issubclass(exc_type, KeyboardInterrupt):
            sys.__excepthook__(exc_type, exc_value, exc_traceback)
            return

        r.error("Uncaught exception", exc_info=(exc_type, exc_value, exc_traceback))

    sys.excepthook = handle_exception


class LoggingMiddleware(BaseHTTPMiddleware):
    """Middleware to ensure requests are logged.

    This middleware will ensure that an access log entry is created. The access
    log will be generated with an event with the text `http request served`.

    The middleware will additionally create a 16byte pseudo-random request
    identifier (represented as a 32byte hex string), attach it to the loggers
    context vars and add it to the outgoing HTTP response in the form of a
    `X-Request-ID` header.

    Attributes:
       app: A FastAPI app.
    """

    def __init__(
        self,
        app,
    ):
        """Initialize the LoggingMiddleware.

        Args:
           app: A FastAPI app.
        """
        super().__init__(app)

    async def dispatch(self, request: Request, call_next):
        """Serve request with the middleware."""
        structlog.contextvars.clear_contextvars()
        request_id = "%032x" % random.randrange(16**32)
        structlog.contextvars.bind_contextvars(request_id=request_id)

        start_time = time.perf_counter_ns()
        response: Response = Response(status_code=418)
        try:
            response = await call_next(request)
        except Exception:
            structlog.stdlib.get_logger().exception("uncaught exception")
            raise
        finally:
            process_time = time.perf_counter_ns() - start_time
            status_code = response.status_code
            response.headers["X-Request-ID"] = request_id
            request.client
            client_port = request.client.port  # type: ignore
            client_host = request.client.host  # type: ignore
            http_method = request.method
            http_version = request.scope["http_version"]
            structlog.stdlib.get_logger().info(
                "http request served",
                http={
                    "url": str(request.url),
                    "status_code": status_code,
                    "method": http_method,
                    "version": http_version,
                },
                network={"client": {"ip": client_host, "port": client_port}},
                duration=process_time,
            )

        return response

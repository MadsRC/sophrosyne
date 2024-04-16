import asyncio
import sys
from contextlib import asynccontextmanager
from functools import wraps

import click
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from sophrosyne.api import api_router
from sophrosyne.core.config import get_settings
from sophrosyne.core.database import (
    create_db_and_tables,
    create_default_profile,
    create_root_user,
    engine,
)
from sophrosyne.core.logging import LoggingMiddleware, get_logger
from sophrosyne.core.security import TLS

log = get_logger()


def async_cmd(func):
    """Decorator to run an async function as a synchronous command.

    This decorator allows you to use async functions as synchronous commands in Click.
    It uses the `asyncio.run()` function to run the async function in a synchronous manner.

    Args:
        func (Callable): The async function to be decorated.

    Returns:
        Callable: The decorated function.

    Example:
        @async_cmd
        async def my_async_command():
            # async code here

        if __name__ == "__main__":
            my_async_command()

    Reference:
        This decorator is based on the solution provided in the following StackOverflow post:
        https://stackoverflow.com/questions/67558717/how-can-i-test-async-click-commands-from-an-async-pytest-function
    """

    @wraps(func)
    def wrapper(*args, **kwargs):
        return asyncio.run(func(*args, **kwargs))

    return wrapper


@asynccontextmanager
async def _lifespan(app: FastAPI):
    """FastAPI lifespan event handler."""
    await create_db_and_tables()
    rt = await create_default_profile()
    if rt:
        log.info("default profile created", profile=rt.name)
    rt = await create_root_user()
    if rt:
        log.info("root user created created", token=rt)
    yield
    log.info("app is shutting down")


@click.command()
def version():
    """Get the version of sophrosyne."""
    from importlib.metadata import version

    print(version("sophrosyne"))


@click.command()
@async_cmd
async def healthcheck():
    """Check the health of the SOPH API service."""
    import sys

    import requests

    # Disable warnings for insecure requests
    requests.packages.urllib3.disable_warnings()

    verify = get_settings().security.outgoing_tls_verify
    if verify and get_settings().security.outgoing_tls_ca_path is not None:
        verify = get_settings().security.outgoing_tls_ca_path

    try:
        resp = requests.get(
            f"https://{get_settings().server.listen_host}:{get_settings().server.port}/health/ping",
            verify=verify,
        )
        if resp.status_code == 200 and resp.text == '"pong"':
            print("API is responding.")
        else:
            print("API returned abnormal response.")
            return sys.exit(1)
    except requests.exceptions.ConnectionError as e:
        # This is not really a nice way of doing this, is there not a better way?
        if "CERTIFICATE_VERIFY_FAILED" in str(e):
            reason = str(e)[str(e).find("certificate verify failed: ") :]
            reason = reason.removeprefix("certificate verify failed: ")
            reason = reason[: reason.rfind(" (")]
            reason = reason.strip()
            print(f"SSL/TLS verification failure: {reason}")
        else:
            print("API is not responding.")
        return sys.exit(1)

    from sqlalchemy.ext.asyncio import async_sessionmaker
    from sqlmodel.ext.asyncio.session import AsyncSession

    from sophrosyne.api.routers.health import do_authenticated_healthcheck

    db_session = async_sessionmaker(
        bind=engine,
        class_=AsyncSession,
        expire_on_commit=False,
    )
    async with db_session() as session:
        hc = await do_authenticated_healthcheck(db_session=session)
        if hc.status == "pass":
            log.info("The server is healthy.")
        else:
            log.info("The server is not healthy.")
            return sys.exit(1)


@click.command()
def run():
    """Run the SOPH API service."""
    try:
        get_settings().security.assert_non_default_cryptographic_material()
    except ValueError as e:
        print(f"configuration error: {e}")
        sys.exit(1)

    app = FastAPI(
        lifespan=_lifespan,
        openapi_url="/.well-known/openapi",
        redoc_url="/docs",
    )

    app.add_middleware(
        CORSMiddleware,
        allow_origins=get_settings().backend_cors_origins,
        allow_credentials=False,
        allow_methods=["*"],
        allow_headers=[],
    )
    app.add_middleware(
        LoggingMiddleware,
    )
    app.include_router(api_router)
    import uvicorn

    tls = TLS(
        certificate_path=get_settings().security.certificate_path,
        key_path=get_settings().security.key_path,
        key_password=get_settings().security.key_password,
    )

    uvicorn.run(
        app,
        host=get_settings().server.listen_host,
        port=get_settings().server.port,
        log_level="info",
        ssl_certfile=tls.to_path(input=tls.certificate),
        ssl_keyfile=tls.to_path(input=tls.private_key),
        # Mypy complains about ssl_keyfile_password being a bytes object, when
        # the argument expects a str. It works because internally in uvicorn,
        # it is passed to the ssl.SSLContext.load_cert_chain() method, which
        # expects a bytes, string or None object. This is probably a bug in
        # uvicorn, but it works as expected.
        ssl_keyfile_password=tls.private_key_password,  # type: ignore
    )

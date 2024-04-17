"""Commands for Sophrosyne."""

#
# Do NOT import any modules from sophrosyne before making sure you've read the
# docstring of the _necessary_evil function.
#

import asyncio
from contextlib import asynccontextmanager
from functools import wraps

import click
from fastapi import FastAPI


def _necessary_evil(path: str):
    """Run some initial setup.

    This code, as the name implies, is a necessary evil to make up for a
    missing feature, or perhaps my own personal shortcomings, of Pydantic. It
    does not seem possible to dynamically specify external configuration files
    via `model_config` in Pydantic, forcing us to have the value of the
    `yaml_file` argument be a variable that is set at module import time. This
    unfortunately creates the side effect that the location of the yaml file
    must be known the config module is imported.

    The way this is handled is to have this function take care of setting the
    necessary environment variables to configure the config module before
    importing it.

    It is imperative that this function is run before any other modules from
    sophrosyne is imported. This is because many other modules import the
    config module, and if that happens before this function is run, everything
    breaks.

    Additionally, because this function is run early and by pretty much all
    commands, it is also used to centralize other things such as initialization
    of logging.
    """
    import os

    if "SOPH__CONFIG_YAML_FILE" not in os.environ:
        os.environ["SOPH__CONFIG_YAML_FILE"] = path
    import sophrosyne.core.config  # noqa: F401
    from sophrosyne.core.config import get_settings
    from sophrosyne.core.logging import initialize_logging

    initialize_logging(
        log_level=get_settings().logging.level_as_int,
        format=get_settings().logging.format,
        event_field=get_settings().logging.event_field,
    )


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
    from sophrosyne.core.logging import get_logger

    log = get_logger()
    from sophrosyne.core.database import (
        create_db_and_tables,
        create_default_profile,
        create_root_user,
    )

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
@click.option("--config", default="config.yaml", help="path to configuration file.")
@click.option(
    "--pretty",
    is_flag=True,
    default=False,
    help="If set, prints configuration with indents for easier reading.",
)
def config(config, pretty):
    """Print the configuration to stdout as JSON."""
    _necessary_evil(config)

    from sophrosyne.core.config import get_settings

    indent = None
    if pretty:
        indent = 2

    print(get_settings().model_dump_json(indent=indent))


@click.command()
@click.option("--config", default="config.yaml", help="path to configuration file.")
@async_cmd
async def healthcheck(config):
    """Check the health of the SOPH API service."""
    _necessary_evil(config)

    import sys

    import requests

    from sophrosyne.core.config import get_settings
    from sophrosyne.core.database import (
        engine,
    )
    from sophrosyne.core.logging import get_logger

    log = get_logger()

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
        if resp.status_code != 200 and resp.text != '"pong"':
            log.error("API returned abnormal response.")
            return sys.exit(1)
    except requests.exceptions.ConnectionError as e:
        # This is not really a nice way of doing this, is there not a better way?
        if "CERTIFICATE_VERIFY_FAILED" in str(e):
            reason = str(e)[str(e).find("certificate verify failed: ") :]
            reason = reason.removeprefix("certificate verify failed: ")
            reason = reason[: reason.rfind(" (")]
            reason = reason.strip()
            log.error(f"SSL/TLS verification failure: {reason}")
        else:
            log.error("API is not responding.")
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
            log.error("The server is not healthy.")
            return sys.exit(1)


@click.command()
@click.option("--config", default="config.yaml", help="path to configuration file.")
def run(config):
    """Run the SOPH API service."""
    _necessary_evil(config)

    import sys

    import uvicorn
    from fastapi import FastAPI
    from fastapi.middleware.cors import CORSMiddleware

    from sophrosyne.api import api_router
    from sophrosyne.core.config import get_settings
    from sophrosyne.core.logging import LoggingMiddleware
    from sophrosyne.core.security import TLS

    try:
        get_settings().security.assert_non_default_cryptographic_material()
    except ValueError as e:
        print(f"configuration error: {e}")
        sys.exit(1)

    tls = TLS(
        certificate_path=get_settings().security.certificate_path,
        key_path=get_settings().security.key_path,
        key_password=get_settings().security.key_password,
    )

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

    uvicorn.run(
        app,
        host=get_settings().server.listen_host,
        port=get_settings().server.port,
        log_level="info",
        log_config=None,
        access_log=False,
        ssl_certfile=tls.to_path(input=tls.certificate),
        ssl_keyfile=tls.to_path(input=tls.private_key),
        # Mypy complains about ssl_keyfile_password being a bytes object, when
        # the argument expects a str. It works because internally in uvicorn,
        # it is passed to the ssl.SSLContext.load_cert_chain() method, which
        # expects a bytes, string or None object. This is probably a bug in
        # uvicorn, but it works as expected.
        ssl_keyfile_password=tls.private_key_password,  # type: ignore
    )

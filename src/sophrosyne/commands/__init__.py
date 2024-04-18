"""Commands for Sophrosyne."""

#
# Do NOT import any modules from sophrosyne before making sure you've read the
# docstring of the necessary_evil function.
#

from contextlib import asynccontextmanager

import click
from fastapi import FastAPI

from sophrosyne.commands import database
from sophrosyne.commands.internal import necessary_evil
from sophrosyne.core import async_cmd


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
    necessary_evil(config)

    from sophrosyne.core.config import get_settings

    indent = None
    if pretty:
        indent = 2

    click.echo(get_settings().model_dump_json(indent=indent))


@click.command()
@click.option("--config", default="config.yaml", help="path to configuration file.")
@async_cmd
async def healthcheck(config):
    """Check the health of the sophrosyne API service."""
    necessary_evil(config)

    import sys

    import requests

    from sophrosyne.core.config import get_settings
    from sophrosyne.core.database import (
        engine,
    )

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
            click.echo("API returned abnormal response.")
            return sys.exit(1)
    except requests.exceptions.ConnectionError as e:
        # This is not really a nice way of doing this, is there not a better way?
        if "CERTIFICATE_VERIFY_FAILED" in str(e):
            reason = str(e)[str(e).find("certificate verify failed: ") :]
            reason = reason.removeprefix("certificate verify failed: ")
            reason = reason[: reason.rfind(" (")]
            reason = reason.strip()
            click.echo(f"SSL/TLS verification failure: {reason}")
        else:
            click.echo("API is not responding.")
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
            click.echo("The server is healthy.")
        else:
            click.echo("The server is not healthy.")
            return sys.exit(1)


@click.command()
@click.option("--config", default="config.yaml", help="path to configuration file.")
def run(config):
    """Run the sophrosyne API service."""
    necessary_evil(config)

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


@click.group()
def root():
    """Sophrosyne - A content moderation API."""
    pass


def setup_and_run_commands():
    """Setup the CLI commands and execute the root command."""
    root.add_command(version)
    root.add_command(run)
    root.add_command(healthcheck)
    root.add_command(config)
    root.add_command(database.cmd)
    root()

# -*- coding: utf-8 -*-
"""Main module for the SOPH API.

This module is the entry point for the SOPH API service. It creates the FastAPI
application and starts the uvicorn server.

Example:
    To start the SOPH API service, run the following command:
        $ python -m sophrosyne.main run

    It's also to run use a filepath to the main module:
        $ python src/sophrosyne/main.py run
"""

import sys

# Remove local directory from sys.path to avoid importing local modules by mistake
# instead of the installed ones. Currently, if this is not in place, the local
# `grpc` module will be imported instead of the installed `grpc` module.
sys.path = sys.path[1:]

import click

from sophrosyne.commands import healthcheck, run, version
from sophrosyne.core.config import get_settings
from sophrosyne.core.logging import get_logger, initialize_logging

log = get_logger()
initialize_logging(
    log_level=get_settings().logging.level_as_int,
    format=get_settings().logging.format,
    event_field=get_settings().logging.event_field,
)


@click.group()
def _cli():
    pass


_cli.add_command(version)
_cli.add_command(run)
_cli.add_command(healthcheck)

if __name__ == "__main__":
    _cli()

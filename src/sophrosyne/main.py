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

#
#
# Do NOT import any modules from sophrosyne, except the commands module, before
# you've read the docstring for the sophrosyne.commands._necessary_evil
# function.
#
#

import sys

# Remove local directory from sys.path to avoid importing local modules by mistake
# instead of the installed ones. Currently, if this is not in place, the local
# `grpc` module will be imported instead of the installed `grpc` module.
sys.path = sys.path[1:]

import click

from sophrosyne.commands import config, healthcheck, migrate, run, version


@click.group()
def _cli():  # NOSONAR
    pass


_cli.add_command(version)
_cli.add_command(run)
_cli.add_command(healthcheck)
_cli.add_command(config)
_cli.add_command(migrate.cmd)

if __name__ == "__main__":
    _cli()

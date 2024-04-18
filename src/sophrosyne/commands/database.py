"""Migration commands for the underlying database."""

import click

from sophrosyne.commands.internal import necessary_evil
from sophrosyne.core import async_cmd

_CONFIG_HELP_TEXT = "path to configuration file."


@click.group(name="database")
def cmd():
    """Database management related commands."""
    pass


@cmd.command()
@click.option(
    "--revision",
    required=True,
    help='The ID of the revision you\'d like to upgrade to. Use "head" to upgrade to the latest.',
    confirmation_prompt=True,
)
@click.option("--config", default="config.yaml", help=_CONFIG_HELP_TEXT)
@async_cmd
async def upgrade(config, revision: str):
    """Update the database."""
    necessary_evil(config)
    from sophrosyne.core.database import upgrade

    await upgrade(revision=revision)


@cmd.command()
@click.option(
    "--revision",
    required=True,
    help='The Id of the revision you\'d like to downgrade to. Use "base" to completely wipe the database.',
    confirmation_prompt=True,
)
@click.option("--config", default="config.yaml", help=_CONFIG_HELP_TEXT)
@async_cmd
async def downgrade(config, revision: str):
    """Downgrade the database."""
    necessary_evil(config)
    from sophrosyne.core.database import downgrade

    await downgrade(revision=revision)


@click.group(name="show")
def show():
    """Commands to read metadata from the database."""
    pass


@show.command()
@click.option("--verbose", default=False, is_flag=True)
@click.option("--config", default="config.yaml", help=_CONFIG_HELP_TEXT)
@async_cmd
async def history(config, verbose: bool):
    """Show the migration history of the database."""
    necessary_evil(config)
    from sophrosyne.core.database import history

    await history(verbose=verbose)


@show.command()
@click.option("--verbose", default=False, is_flag=True)
@click.option("--config", default="config.yaml", help=_CONFIG_HELP_TEXT)
@async_cmd
async def current(config, verbose: bool):
    """Show the current migration revision of the database."""
    necessary_evil(config)
    from sophrosyne.core.database import current

    await current(verbose=verbose)


cmd.add_command(show)
cmd.add_command(upgrade)
cmd.add_command(downgrade)

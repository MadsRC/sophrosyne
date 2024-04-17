"""Migration commands for the underlying database."""

import asyncio
from functools import wraps

import click


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


@click.group(name="migrate")
@click.option("--config", default="config.yaml", help="path to configuration file.")
def cmd(config):
    """Run database migrations."""
    pass


@cmd.command()
@click.option("--revision", required=True)
@async_cmd
async def upgrade(revision: str):
    """Update the database."""
    from sophrosyne.core.database import upgrade

    await upgrade(revision=revision)


@cmd.command()
@click.option("--revision", required=True)
@async_cmd
async def downgrade(revision: str):
    """Downgrade the database."""
    from sophrosyne.core.database import downgrade

    await downgrade(revision=revision)


@click.group(name="show")
@click.option("--config", default="config.yaml", help="path to configuration file.")
def show(config):
    """Run database migrations."""
    pass


@show.command()
@click.option("--verbose", default=False, is_flag=True)
@async_cmd
async def history(verbose: bool):
    from sophrosyne.core.database import history

    await history(verbose=verbose)


@show.command()
@click.option("--verbose", default=False, is_flag=True)
@async_cmd
async def current(verbose: bool):
    from sophrosyne.core.database import current

    await current(verbose=verbose)


cmd.add_command(show)
cmd.add_command(upgrade)
cmd.add_command(downgrade)

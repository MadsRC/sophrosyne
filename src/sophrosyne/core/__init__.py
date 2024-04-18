"""Core module for the SOPH API.

This module contains the core logic of the SOPH API service. It defines the
database operations, configuration, and utility functions not tied to a
specific API version.
"""

import asyncio
from functools import wraps


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

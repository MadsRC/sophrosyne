"""Dependencies for the API."""

from typing import Annotated

from fastapi import Depends, HTTPException
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from sqlalchemy.ext.asyncio import async_sessionmaker
from sqlmodel import select
from sqlmodel.ext.asyncio.session import AsyncSession

from sophrosyne.core.config import Settings
from sophrosyne.core.config import get_settings as get_config_settings
from sophrosyne.core.database import engine
from sophrosyne.core.logging import get_logger
from sophrosyne.core.models import User
from sophrosyne.core.safety import Safety
from sophrosyne.core.security import sign

header_scheme = HTTPBearer(
    description="Authorization header for the API",
    auto_error=False,
)


async def get_db_session():
    """Returns an asynchronous database session.

    This function creates an asynchronous session using the `async_sessionmaker` and `AsyncSession` classes.
    The session is yielded to the caller, allowing them to use it within a context manager.

    Returns:
        AsyncSession: An asynchronous database session.

    Example:
        async with get_db_session() as session:
            # Use the session to perform database operations
            await session.execute(...)
            await session.commit()
    """
    async_session = async_sessionmaker(
        engine, class_=AsyncSession, expire_on_commit=False
    )
    async with async_session() as session:
        yield session


async def get_safety_service(
    db_session: Annotated[AsyncSession, Depends(get_db_session)],
):
    """Retrieves the safety service.

    Args:
        db_session (AsyncSession): The database session.

    Yields:
        Safety: The safety service.

    """
    yield Safety(db_session=db_session)


async def _authenticate(
    credentials: Annotated[HTTPAuthorizationCredentials, Depends(header_scheme)],
    db_session: Annotated[AsyncSession, Depends(get_db_session)],
) -> User | None:
    """Authenticates a request based on provided credentials.

    If no credentials are provided, None is returned.

    Args:
       credentials (HTTPAuthorizationCredentials): The credentials to check.
       db_session (AsyncSession): The database session to use.

    Returns:
        User | None: The authenticated user. If credentials matches no user,
        None is returned
    """
    if credentials is None:
        get_logger().info("authentication", result="fail")
        return None
    result = await db_session.exec(
        select(User).where(User.signed_token == sign(credentials.credentials))
    )
    user = result.first()
    if user is None:
        get_logger().info("authentication", result="fail")
        return None
    get_logger().info("authentication", result="success", user=user.id)
    return user


async def is_authenticated(
    user: Annotated[User | None, Depends(_authenticate)],
) -> bool:
    """Checks that the user is authenticated and returns a boolean.

    Args:
        user (User | None): The user to check.

    Returns:
        bool: True if the user is authenticated, False otherwise.
    """
    return user is not None


async def auth_and_return_user(
    user: Annotated[User | None, Depends(_authenticate)],
) -> User:
    """Retrieves the current user based on the provided credentials.

    Args:
        user (User): The user to authenticate.

    Returns:
        User: The current user.

    Raises:
        HTTPException: If the user is not authenticated (status code 403).
    """
    if user is None:
        raise HTTPException(status_code=403, detail="Not authenticated")

    return user


async def get_settings() -> Settings:
    """Retrieves the application settings.

    Returns:
        Settings: The application settings.
    """
    return get_config_settings()


async def require_admin(
    current_user: Annotated[User, Depends(auth_and_return_user)],
):
    """Requires the current user to be an admin.

    Args:
        current_user (User): The current user.

    Raises:
        HTTPException: If the current user is not an admin (status code 403).
    """
    if not current_user.is_admin:
        get_logger().info("user is not an admin", user=current_user.id)
        raise HTTPException(
            status_code=403, detail="Only admins can perform this operation"
        )


async def require_active_user(
    current_user: Annotated[User, Depends(auth_and_return_user)],
):
    """Requires the current user to be active.

    Args:
        current_user (User): The current user.

    Raises:
        HTTPException: If the current user is not active (status code 403).
    """
    if not current_user.is_active:
        get_logger().info("user is not active", user=current_user.id)
        raise HTTPException(status_code=403, detail="Not authenticated")

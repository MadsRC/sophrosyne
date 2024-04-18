"""Users API endpoints.

Attributes:
    router (APIRouter): The FastAPI router for the users API.
"""

from fastapi import APIRouter, Depends, HTTPException, Query
from sqlmodel import select
from sqlmodel.ext.asyncio.session import AsyncSession

from sophrosyne.api.dependencies import (
    auth_and_return_user,
    get_db_session,
    require_admin,
)
from sophrosyne.api.v1.models import (
    UsersCreateUserRequest,
    UsersCreateUserResponse,
    UsersDeleteUserRequest,
    UsersDeleteUserResponse,
    UsersListUserRequest,
    UsersListUserResponse,
    UsersListUsersResponse,
    UsersRotateTokenRequest,
    UsersRotateTokenResponse,
    UsersUpdateUserRequest,
    UsersUpdateUserResponse,
)
from sophrosyne.api.v1.tags import Tags
from sophrosyne.core.models import User
from sophrosyne.core.security import new_token, sign

router = APIRouter()

USER_NOT_FOUND = "User not found"


@router.post(
    "/users/create-user",
    response_model=UsersCreateUserResponse,
    tags=[Tags.users],
    dependencies=[Depends(require_admin)],
)
async def create_user(
    *,
    db_session: AsyncSession = Depends(get_db_session),
    user: UsersCreateUserRequest,
):
    """Create a new user.

    Args:
        db_session (AsyncSession): The database session.
        user (UsersCreateUserRequest): The request payload containing user data.

    Returns:
        UsersCreateUserResponse: The newly created user.

    Raises:
        None

    """
    token = new_token()
    extra_data = {"signed_token": sign(token)}
    db_user = User.model_validate(user, update=extra_data)
    db_session.add(db_user)
    await db_session.commit()
    await db_session.refresh(db_user)
    return UsersCreateUserResponse.model_validate(db_user, update={"token": token})


@router.get(
    "/users/list-users",
    response_model=UsersListUsersResponse,
    tags=[Tags.users],
    dependencies=Depends(require_admin),
)
async def read_users(
    *,
    db_session: AsyncSession = Depends(get_db_session),
    offset: int = 0,
    limit: int = Query(100, le=100),
):
    """Retrieve a list of users from the database.

    Args:
        db_session (AsyncSession): The database session.
        offset (int): The offset for pagination. Defaults to 0.
        limit (int): The maximum number of users to retrieve. Defaults to 100.

    Returns:
        UsersListUsersResponse: A list of user objects.

    Raises:
        HTTPException: If no users are found in the database.
    """
    result = await db_session.exec(select(User).offset(offset).limit(limit))
    users = result.all()
    if not users:
        raise HTTPException(status_code=400, detail="No users found")
    return users


@router.post(
    "/users/list-user", response_model=UsersListUserResponse, tags=[Tags.users]
)
async def read_user(
    *,
    db_session: AsyncSession = Depends(get_db_session),
    req: UsersListUserRequest,
    current_user=Depends(auth_and_return_user),
):
    """Retrieve a user from the database based on the provided name.

    Args:
        db_session (AsyncSession): The database session.
        req (UsersListUserRequest): The request object containing the user's name.
        current_user (User): The current user making the request.

    Returns:
        UsersListUserResponse: The user object retrieved from the database.

    Raises:
        HTTPException: If the user is not found in the database.
    """
    result = await db_session.exec(select(User).where(User.name == req.name))
    user = result.first()
    if not user:
        raise HTTPException(status_code=400, detail=USER_NOT_FOUND)
    if current_user.name != user.name:
        raise HTTPException(status_code=403, detail=USER_NOT_FOUND)
    return user


@router.patch(
    "/users/update-user",
    response_model=UsersUpdateUserResponse,
    tags=[Tags.users],
    dependencies=[Depends(require_admin)],
)
async def update_user(
    *,
    db_session: AsyncSession = Depends(get_db_session),
    req: UsersUpdateUserRequest,
):
    """Update a user in the database.

    Args:
        db_session: The database session.
        req: The request object containing the updated user data.

    Returns:
        UsersUpdateUserResponse: The updated user object.

    Raises:
        HTTPException: If the user is not found in the database.
    """
    result = await db_session.exec(select(User).where(User.name == req.name))
    db_user = result.first()
    if not db_user:
        raise HTTPException(status_code=400, detail=USER_NOT_FOUND)
    user_data = req.model_dump(exclude_unset=True)
    db_user.sqlmodel_update(user_data)
    db_session.add(db_user)
    await db_session.commit()
    await db_session.refresh(db_user)
    return db_user


@router.delete(
    "/users/delete-user",
    response_model=UsersDeleteUserResponse,
    tags=[Tags.users],
    dependencies=[Depends(require_admin)],
)
async def delete_user(
    *,
    db_session: AsyncSession = Depends(get_db_session),
    req: UsersDeleteUserRequest,
):
    """Delete a user from the database.

    Args:
        db_session (AsyncSession): The database session.
        req (UsersDeleteUserRequest): The request object containing the user's name.

    Returns:
        UsersDeleteUserResponse: The response object indicating the success of the operation.
    """
    result = await db_session.exec(select(User).where(User.name == req.name))
    db_user = result.first()
    if not db_user:
        raise HTTPException(status_code=400, detail=USER_NOT_FOUND)
    await db_session.delete(db_user)
    await db_session.commit()
    return UsersDeleteUserResponse(ok=True)


@router.post(
    "/users/rotate-token", response_model=UsersRotateTokenResponse, tags=[Tags.users]
)
async def rotate_user_token(
    *,
    db_session: AsyncSession = Depends(get_db_session),
    req: UsersRotateTokenRequest,
    current_user=Depends(auth_and_return_user),
):
    """Create a new token for a user and update it in the database.

    Args:
        db_session (AsyncSession): The database session.
        req (UsersRotateTokenRequest): The request object containing the user's name.
        current_user (User): The current user making the request.

    Returns:
        UsersRotateTokenResponse: The updated user object with the new token.

    Raises:
        HTTPException: If the user is not found in the database.
    """
    result = await db_session.exec(select(User).where(User.name == req.name))
    db_user = result.first()
    if not db_user:
        raise HTTPException(status_code=400, detail=USER_NOT_FOUND)
    if current_user.name != db_user.name and not current_user.is_admin:
        raise HTTPException(status_code=403, detail="Not authorized")
    token = new_token()
    db_user.signed_token = sign(token)
    db_session.add(db_user)
    await db_session.commit()
    await db_session.refresh(db_user)
    return UsersRotateTokenResponse.model_validate(db_user, update={"token": token})

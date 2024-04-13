"""Profiles API endpoints.

Attributes:
    router (APIRouter): The FastAPI router for the profiles API.
"""

from typing import Sequence

from fastapi import APIRouter, Depends, HTTPException, Query
from sqlmodel import col, select
from sqlmodel.ext.asyncio.session import AsyncSession

from sophrosyne.api.dependencies import get_db_session, require_admin
from sophrosyne.api.v1 import Tags
from sophrosyne.api.v1.models import (
    ProfilesCreateProfileRequest,
    ProfilesCreateProfileResponse,
    ProfilesDeleteProfileRequest,
    ProfilesDeleteProfileResponse,
    ProfilesListProfileRequest,
    ProfilesListProfileResponse,
    ProfilesListProfilesResponse,
    ProfilesUpdateProfileRequest,
    ProfilesUpdateProfileResponse,
)
from sophrosyne.core.models import Check, Profile

router = APIRouter(dependencies=[Depends(require_admin)])


@router.post(
    "/profiles/create-profile",
    response_model=ProfilesCreateProfileResponse,
    tags=[Tags.profiles],
)
async def create_profile(
    *,
    db_session: AsyncSession = Depends(get_db_session),
    req: ProfilesCreateProfileRequest,
):
    """Create a new profile in the database.

    Args:
        db_session (AsyncSession): The database session.
        req (ProfilesCreateProfileRequest): The request object containing the profile data.

    Returns:
        The created profile object.

    Raises:
        None.
    """
    if req.checks is not None:
        result = await db_session.exec(
            select(Check).where(col(Check.name).in_(req.checks))
        )
        db_checks = result.all()
        # Clear the profiles to avoid "'int' object has no attribute '_sa_instance_state'"
        # error when converting to our SQLModel table model later.
        req.checks.clear()

    db_profile = Profile.model_validate(req)
    db_profile.checks.extend(db_checks)
    db_session.add(db_profile)
    await db_session.commit()
    await db_session.refresh(db_profile)
    return ProfilesCreateProfileResponse.model_validate(
        db_profile, update={"checks": [c.name for c in db_profile.checks]}
    )


@router.get(
    "/profiles/list-profiles",
    response_model=ProfilesListProfilesResponse,
    tags=[Tags.profiles],
)
async def read_profiles(
    *,
    db_session: AsyncSession = Depends(get_db_session),
    offset: int = 0,
    limit: int = Query(100, le=100),
):
    """Retrieve a list of profiles with pagination support.

    Args:
        db_session (AsyncSession): The database session.
        offset (int): The offset for pagination. Defaults to 0.
        limit (int): The maximum number of profiles to retrieve. Defaults to 100.

    Returns:
        List[Profile]: A list of profiles matching the query.

    Raises:
        HTTPException: If no profiles are found.
    """
    result = await db_session.exec(select(Profile).offset(offset).limit(limit))
    profiles = result.all()
    if not profiles:
        raise HTTPException(status_code=404, detail="No profiles found")
    return profiles


@router.post(
    "/profiles/list-profile",
    response_model=ProfilesListProfileResponse,
    tags=[Tags.profiles],
)
async def read_profile(
    *,
    db_session: AsyncSession = Depends(get_db_session),
    req: ProfilesListProfileRequest,
):
    """Retrieve a profile from the database based on the provided name.

    Args:
        db_session (AsyncSession): The database session.
        req (ProfilesListProfileRequest): The request object containing the profile name.

    Returns:
        Profile: The retrieved profile.

    Raises:
        HTTPException: If the profile is not found in the database.
    """
    result = await db_session.exec(select(Profile).where(Profile.name == req.name))
    profile = result.first()
    if not profile:
        raise HTTPException(status_code=404, detail="Profile not found")
    return ProfilesListProfileResponse.model_validate(
        profile, update={"checks": [c.name for c in profile.checks]}
    )


@router.patch(
    "/profiles/update-profile",
    response_model=ProfilesUpdateProfileResponse,
    tags=[Tags.profiles],
)
async def update_profile(
    *,
    db_session: AsyncSession = Depends(get_db_session),
    req: ProfilesUpdateProfileRequest,
):
    """Update a profile in the database.

    Args:
        db_session (AsyncSession): The database session.
        req (ProfilesUpdateProfileRequest): The request object containing the updated profile data.

    Returns:
        The updated profile object.

    Raises:
        HTTPException: If the profile is not found in the database.
    """
    result = await db_session.exec(select(Profile).where(Profile.name == req.name))
    db_profile = result.first()
    if not db_profile:
        raise HTTPException(status_code=404, detail="Profile not found")

    if req.checks is not None:
        db_checks = await db_session.exec(
            select(Check).where(col(Check.name).in_(req.checks))
        )
        db_profile.checks.clear()
        db_profile.checks.extend(db_checks)

    profile_data = req.model_dump(exclude_unset=True)
    db_profile.sqlmodel_update(profile_data)
    db_session.add(db_profile)
    await db_session.commit()
    await db_session.refresh(db_profile)
    return db_profile


@router.delete(
    "/profiles/delete-profile",
    response_model=ProfilesDeleteProfileResponse,
    tags=[Tags.profiles],
)
async def delete_profile(
    *,
    db_session: AsyncSession = Depends(get_db_session),
    req: ProfilesDeleteProfileRequest,
):
    """Delete a profile from the database.

    Args:
        db_session (AsyncSession): The database session.
        req (ProfilesDeleteProfileRequest): The request object containing the profile name.

    Returns:
        ProfilesDeleteProfileResponse: The response object indicating the success of the operation.

    Raises:
        HTTPException: If the profile is not found in the database.
    """
    result = await db_session.exec(select(Profile).where(Profile.name == req.name))
    db_profile = result.first()
    if not db_profile:
        raise HTTPException(status_code=404, detail="Profile not found")
    await db_session.delete(db_profile)
    await db_session.commit()
    return ProfilesDeleteProfileResponse(ok=True)

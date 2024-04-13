"""API endpoints for managing checks.

Attributes:
    router (APIRouter): The FastAPI router for the checks endpoints.
"""

from typing import Sequence

from fastapi import APIRouter, Depends, HTTPException, Query
from sqlmodel import col, select
from sqlmodel.ext.asyncio.session import AsyncSession

from sophrosyne.api.dependencies import get_db_session, require_admin
from sophrosyne.api.v1.models import (
    ChecksCreateCheckRequest,
    ChecksCreateCheckResponse,
    ChecksDeleteCheckRequest,
    ChecksDeleteCheckResponse,
    ChecksListCheckRequest,
    ChecksListCheckResponse,
    ChecksListChecksResponse,
    ChecksUpdateCheckRequest,
    ChecksUpdateCheckResponse,
)
from sophrosyne.api.v1.tags import Tags
from sophrosyne.core.models import Check, Profile

router = APIRouter(dependencies=[Depends(require_admin)])


@router.post(
    "/checks/create-check", response_model=ChecksCreateCheckResponse, tags=[Tags.checks]
)
async def create_check(
    *, db_session: AsyncSession = Depends(get_db_session), req: ChecksCreateCheckRequest
):
    """Create a new check.

    Args:
        db_session (AsyncSession): The database session.
        req (ChecksCreateCheckRequest): The request object containing the check details.

    Returns:
        ChecksCreateCheckResponse: The response object containing the created check.
    """
    db_profiles: Sequence[Profile] = []
    if req.profiles is not None:
        result = await db_session.exec(
            select(Profile).where(col(Profile.name).in_(req.profiles))
        )
        db_profiles = result.all()
        # Clear the profiles to avoid "'int' object has no attribute '_sa_instance_state'"
        # error when converting to our SQLModel table model later.
        req.profiles.clear()

    db_check = Check.model_validate(req)
    db_check.profiles.extend(db_profiles)
    db_session.add(db_check)
    await db_session.commit()
    await db_session.refresh(db_check)
    return ChecksCreateCheckResponse.model_validate(
        db_check, update={"profiles": [p.name for p in db_check.profiles]}
    )


@router.get(
    "/checks/list-checks", response_model=ChecksListChecksResponse, tags=[Tags.checks]
)
async def read_checks(
    *,
    db_session: AsyncSession = Depends(get_db_session),
    offset: int = 0,
    limit: int = Query(100, le=100),
):
    """Retrieve a list of checks.

    Args:
        db_session (AsyncSession): The database session.
        offset (int): The offset for pagination.
        limit (int): The maximum number of checks to retrieve.

    Returns:
        List[Check]: A list of checks.

    Raises:
        HTTPException: If no checks are found.
    """
    result = await db_session.exec(select(Check).offset(offset).limit(limit))
    checks = result.all()
    if not checks:
        raise HTTPException(status_code=404, detail="No checks found")
    return checks


@router.post(
    "/checks/list-check", response_model=ChecksListCheckResponse, tags=[Tags.checks]
)
async def read_check(
    *, db_session: AsyncSession = Depends(get_db_session), req: ChecksListCheckRequest
):
    """Retrieve a specific check by its name.

    Args:
        db_session (AsyncSession): The database session.
        req (ChecksListCheckRequest): The request object containing the name of the check.

    Returns:
        Check: The check object if found.

    Raises:
        HTTPException: If the check is not found (status code 404).
    """
    result = await db_session.exec(select(Check).where(Check.name == req.name))
    check = result.first()
    if not check:
        raise HTTPException(status_code=404, detail="Check not found")
    return ChecksListCheckResponse.model_validate(
        check, update={"profiles": [p.name for p in check.profiles]}
    )


@router.patch(
    "/checks/update-check", response_model=ChecksUpdateCheckResponse, tags=[Tags.checks]
)
async def update_check(
    *,
    db_session: AsyncSession = Depends(get_db_session),
    req: ChecksUpdateCheckRequest,
):
    """Update a check in the database.

    Args:
        db_session (AsyncSession): The database session.
        req (ChecksUpdateCheckRequest): The request object containing the updated check data.

    Returns:
        The updated check object.

    Raises:
        HTTPException: If the check is not found in the database.
    """
    result = await db_session.exec(select(Check).where(Check.name == req.name))
    db_check = result.first()
    if not db_check:
        raise HTTPException(status_code=404, detail="Check not found")

    if req.profiles is not None:
        db_profiles = await db_session.exec(
            select(Profile).where(col(Profile.name).in_(req.profiles))
        )
        db_check.profiles.clear()
        db_check.profiles.extend(db_profiles)

    check_data = req.model_dump(exclude_unset=True)
    db_check.sqlmodel_update(check_data)
    db_session.add(db_check)
    await db_session.commit()
    await db_session.refresh(db_check)
    return db_check


@router.delete("/checks/delete-check", tags=[Tags.checks])
async def delete_check(
    *, db_session: AsyncSession = Depends(get_db_session), req: ChecksDeleteCheckRequest
):
    """Delete a check from the database.

    Args:
        db_session (AsyncSession): The database session.
        req (ChecksDeleteCheckRequest): The request object containing the name of the check to delete.

    Returns:
        ChecksDeleteCheckResponse: The response object indicating the success of the deletion.
    """
    result = await db_session.exec(select(Check).where(Check.name == req.name))
    db_check = result.first()
    if not db_check:
        raise HTTPException(status_code=404, detail="Check not found")
    await db_session.delete(db_check)
    await db_session.commit()
    return ChecksDeleteCheckResponse(ok=True)

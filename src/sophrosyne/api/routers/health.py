"""This module contains the health router for the application."""

import importlib.metadata
import time
from datetime import datetime, timezone
from typing import Annotated

from fastapi import APIRouter, Depends, Response, status
from sophrosyne.api.dependencies import get_db_session, get_settings, is_authenticated
from sophrosyne.core.config import Settings
from sophrosyne.healthcheck.models import Check, HealthCheck, Status, SubComponent
from sqlmodel import text
from sqlmodel.ext.asyncio.session import AsyncSession

router = APIRouter()


@router.get("", response_model=HealthCheck)
async def get_health(
    response: Response,
    settings: Annotated[Settings, Depends(get_settings)],
    db_session: Annotated[AsyncSession, Depends(get_db_session)],
    is_authenticated: Annotated[bool, Depends(is_authenticated)],
) -> HealthCheck:
    """Get the health of the application.

    Args:
        response (Response): The response object.
        settings (Settings): The application settings.
        db_session (AsyncSession): The database session.
        is_authenticated (bool): Flag indicating if the user is authenticated.

    Returns:
        HealthCheck: The health of the application.
    """
    response.headers["Cache-Control"] = "no-cache"
    hc: HealthCheck
    if is_authenticated:
        hc = await do_authenticated_healthcheck(db_session=db_session)
    else:
        hc = await do_healthcheck(db_session=db_session)

    if hc.status == Status.PASS:
        response.status_code = status.HTTP_200_OK
    else:
        response.status_code = status.HTTP_503_SERVICE_UNAVAILABLE
    return hc


@router.get("/ping", response_model=str)
async def ping() -> str:
    """Ping endpoint to check if the server is running."""
    return "pong"


async def do_authenticated_healthcheck(db_session: AsyncSession) -> HealthCheck:
    """Perform an authenticated health check.

    Args:
        db_session (AsyncSession): The database session.

    Returns:
        HealthCheck: The health check result.
    """
    status: Status
    output: str | None = None
    end: int
    begin = time.perf_counter_ns()
    try:
        await db_session.execute(statement=text("SELECT 1"))
        end = time.perf_counter_ns()
        status = Status.PASS
    except Exception as e:
        end = time.perf_counter_ns()
        status = Status.FAIL
        output = str(e)

    return HealthCheck(
        status=status,
        version=importlib.metadata.version("sophrosyne"),
        checks=Check(
            sub_components={
                "database:responseTime": [
                    SubComponent(
                        status=status,
                        output=output,
                        observed_value=str(end - begin),
                        observed_unit="ns",
                        component_type="datastore",
                        time=datetime.now(timezone.utc),
                    )
                ],
            }
        ),
    )


async def do_healthcheck(db_session: AsyncSession) -> HealthCheck:
    """Perform a health check.

    Args:
        db_session (AsyncSession): The database session.

    Returns:
        HealthCheck: The health check result.
    """
    try:
        await db_session.execute(statement=text("SELECT 1"))
    except Exception:
        return HealthCheck(
            status=Status.FAIL,
        )
    return HealthCheck(
        status=Status.PASS,
    )

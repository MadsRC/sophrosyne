"""Safety API endpoints.

Attributes:
    router (APIRouter): The FastAPI router for the safety endpoints.
"""

from typing import Annotated

from fastapi import APIRouter, Body, Depends

from sophrosyne.api.dependencies import auth_and_return_user, get_safety_service
from sophrosyne.api.v1.models import (
    SafetyScanRequest,
    SafetyScanResponse,
)
from sophrosyne.core.config import Settings, get_settings
from sophrosyne.core.models import (
    SafetyServicePayload,
    SafetyServicePayloadImage,
    SafetyServicePayloadText,
    User,
)
from sophrosyne.core.safety import Safety

router = APIRouter()


@router.post("/safety/scan", response_model=SafetyScanResponse)
async def safety(
    req: Annotated[SafetyScanRequest, Body()],
    ss: Annotated[Safety, Depends(get_safety_service)],
    current_user: Annotated[User, Depends(auth_and_return_user)],
    settings: Annotated[Settings, Depends(get_settings)],
) -> SafetyScanResponse:
    """Endpoint for performing a safety scan.

    Args:
        req (SafetyScanRequest): The request payload containing the scan details.
        ss (Safety): The safety service dependency.
        current_user (User): The current user making the request.
        settings (Settings): The application settings.

    Returns:
        SafetyScanResponse: The response containing the scan results.
    """
    ssp: SafetyServicePayload
    profile: str
    if current_user.default_profile is None:
        profile = settings.default_profile
    else:
        profile = current_user.default_profile
    if req.kind == "text":
        ssp = SafetyServicePayloadText.model_validate(req.model_dump())
    elif req.kind == "image":
        ssp = SafetyServicePayloadImage.model_validate(req.model_dump())
    ssr = await ss.predict(profile=profile, data=ssp)

    return SafetyScanResponse.model_validate(ssr.model_dump())

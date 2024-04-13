"""Module containing the API classes for the SOPH API.

Attributes:
    api_router (APIRouter): The API router for the API.
"""

from typing import Annotated

from fastapi import APIRouter, Depends

from sophrosyne.api.routers.health import router as health_router
from sophrosyne.api.v1.api import api_router as v1_api_router
from sophrosyne.core.config import get_settings

api_router = APIRouter()
api_router.include_router(v1_api_router, prefix=get_settings().api_v1_str)
api_router.include_router(health_router, prefix="/health")

"""Exposes the API routes for the v1 version of the API.

This module contains the API routes for the v1 version of the API. The routes are
grouped by the different resources they interact with.

Attributes:
    api_router (APIRouter): The API router for the v1 version of the API.
"""


from fastapi import APIRouter, Depends

from sophrosyne.api.dependencies import auth_and_return_user, require_active_user
from sophrosyne.api.v1.routers import checks, profiles, safety, users

api_router = APIRouter(
    dependencies=[Depends(auth_and_return_user), Depends(require_active_user)]
)
api_router.include_router(safety.router)
api_router.include_router(users.router)
api_router.include_router(checks.router)
api_router.include_router(profiles.router)

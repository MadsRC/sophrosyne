"""Tags for the API endpoints."""

from enum import Enum


class Tags(Enum):
    """Tags for the API endpoints.

    Attributes:
        profiles (str): The profiles tag.
        checks (str): The checks tag.
        users (str): The users tag.
        safety (str): The safety tag.
    """

    profiles = "profiles"
    checks = "checks"
    users = "users"
    safety = "safety"

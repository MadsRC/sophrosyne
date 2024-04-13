"""SQLModel, Pydantic models and Pydantic schemas for the SOPH API.

This module defines the SQLModels, Pydantic models and Pydantic schemas for the
SOPH API service. Any model or schema that is used in the core logic of the
application should be defined here. If a model or schema is used to define the
request or response body of an API endpoint, it should be defined in the
models module of the API version it belongs to.

Example:
    The User model is used to define the User table in the database. The
    UserBase model is used to define the base fields of the User model.

Any model or schema that is used in the core logic of the application should be
validated against these models and schemas.
"""

from datetime import datetime
from enum import Enum
from typing import Annotated, Literal, Union

from pydantic import BaseModel, EmailStr
from sqlmodel import (
    ARRAY,
    JSON,
    AutoString,
    Column,
    Field,
    Relationship,
    SQLModel,
    String,
)

from sophrosyne.core.config import get_settings


class SafetyServicePayloadType(str, Enum):
    """Enum class representing the payload types for the Safety Service."""

    TEXT = "text"
    IMAGE = "image"


class UserBase(SQLModel):
    """Base class for all users.

    Attributes:
        name (str): The name of the user.
        contact (EmailStr): The contact email of the user.
        is_active (bool): Whether the user is active or not.
        default_profile (str): The default profile of the user.
    """

    name: str = Field(unique=True, index=True)
    created_at: datetime = Field(default_factory=datetime.utcnow)
    contact: EmailStr = Field(sa_type=AutoString)
    is_active: bool = Field(default=True)
    default_profile: str | None = Field(
        default=get_settings().default_profile, foreign_key="profile.name"
    )


class User(UserBase, table=True):
    """Model for the User table in the database.

    Attributes:
        id (int): The ID of the user.
        signed_token (str): The token of the user, signed with sophrosyne.core.security.sign.
    """

    id: int | None = Field(default=None, primary_key=True)
    signed_token: str = Field(index=True, unique=True)
    is_admin: bool = Field(default=False)


class ProfileCheckAssociation(SQLModel, table=True):
    """Model for the ProfileCheckAssociation table in the database.

    Attributes:
        profile_id (int): The ID of the profile.
        check_id (int): The ID of the check.
    """

    profile_id: int | None = Field(
        default=None, foreign_key="profile.id", primary_key=True
    )
    check_id: int | None = Field(default=None, foreign_key="check.id", primary_key=True)


class CheckBase(SQLModel):
    """Model for the Check table in the database.

    Attributes:
        name (str): The name of the check.
        created_at (datetime): The creation date of the check.
        upstream_services (list[str]): The list of upstream services for the check.
        config (dict[str, Union[str, int, float, bool]]): The configuration for the check.
    """

    name: str = Field(unique=True, index=True)
    created_at: datetime = Field(default_factory=datetime.utcnow)
    upstream_services: list[str] = Field(
        default_factory=list, sa_column=Column(ARRAY(String))
    )
    supported_types: list[SafetyServicePayloadType] = Field(
        default_factory=list, sa_column=Column(ARRAY(String))
    )
    config: dict[str, Union[str, int, float, bool]] = Field(
        default_factory=dict, sa_column=Column(JSON)
    )


class Check(CheckBase, table=True):
    """Model for the Check table in the database.

    Attributes:
        id (int): The ID of the check.
        profiles (list[Profile]): The profiles that use the check.
    """

    id: int | None = Field(default=None, primary_key=True)
    profiles: list["Profile"] = Relationship(
        back_populates="checks",
        link_model=ProfileCheckAssociation,
        # Prevent SQLAlchemy from lazy loading the relationship. See https://stackoverflow.com/questions/74252768/missinggreenlet-greenlet-spawn-has-not-been-called
        sa_relationship_kwargs={"lazy": "selectin"},
    )


class ProfileBase(SQLModel):
    """Model for the Profile table in the database.

    Attributes:
        name (str): The name of the profile.
        created_at (datetime): The creation date of the profile.
    """

    name: str = Field(unique=True, index=True)
    created_at: datetime = Field(default_factory=datetime.utcnow)


class Profile(ProfileBase, table=True):
    """Represents a profile in the system.

    Attributes:
        id (int | None): The ID of the profile. Defaults to None.
        checks (list["Check"]): The list of checks associated with the profile.
    """

    id: int | None = Field(default=None, primary_key=True)
    checks: list["Check"] = Relationship(
        back_populates="profiles",
        link_model=ProfileCheckAssociation,
        # Prevent SQLAlchemy from lazy loading the relationship. See https://stackoverflow.com/questions/74252768/missinggreenlet-greenlet-spawn-has-not-been-called
        sa_relationship_kwargs={"lazy": "selectin"},
    )


class SafetyServicePayloadText(BaseModel):
    """Represents the payload for analyzing the safety of a text.

    Attributes:
        text (str): The text to be analyzed for safety.
    """

    text: str = Field(
        title="Text",
        description="The text to be analyzed for safety.",
        min_length=1,
        max_length=1000,
    )


class SafetyServicePayloadImage(BaseModel):
    """Represents the payload for analyzing the safety of an image.

    Attributes:
        image (str): The image to be analyzed for safety.
    """

    image: str = Field(
        title="Image",
        description="The image to be analyzed for safety.",
        min_length=1,
        max_length=1000,
    )


SafetyServicePayload = Union[SafetyServicePayloadText, SafetyServicePayloadImage]


class Verdict(BaseModel):
    """Represents a safety verdict.

    Attributes:
        verdict (bool): The safety verdict.
        checks (dict[str, bool]): The safety checks.
    """

    verdict: bool = Field(
        title="Verdict",
        description="The safety verdict.",
    )
    checks: dict[str, bool] = Field(
        title="Checks",
        description="The safety checks.",
    )

"""SQLModels, Pydantic models and Pydantic schemas for the v1 API.

This module defines the SQLModels, Pydantic models and Pydantic schemas for the
v1 API of the SOPH API service. The models and schemas are used to define the
request and response bodies for the API endpoints.

Example:
    The UserCreate Pydantic model is used to define the request body for the
    POST /users/ endpoint. The UserBase Pydantic model is used to define the
    response body for the GET /users/ endpoint.

Any input or output to the v1 API should be validated against these models and
schemas.

If a model or schema doesn't directy affect the input or output of the API, it
should be defined in the core.models module instead.
"""

from typing import Annotated, Literal, Union

from pydantic import EmailStr, Field
from sqlmodel import SQLModel

from sophrosyne.core.models import (
    CheckBase,
    ProfileBase,
    SafetyServicePayloadImage,
    SafetyServicePayloadText,
    SafetyServicePayloadType,
    UserBase,
    Verdict,
)


class UsersCreateUserResponse(UserBase):
    """Represents the response object for creating a user.

    Attributes:
        token (str): The authentication token associated with the user.
    """

    token: str


class UsersCreateUserRequest(SQLModel):
    """Represents a request to create a new user.

    Attributes:
        name (str): The name of the user.
        contact (EmailStr): The contact email of the user.
        is_active (bool, optional): Whether the user is active. Defaults to True.
        default_profile (str | None, optional): The default profile of the user. Defaults to None.
    """

    name: str
    contact: EmailStr
    is_active: bool = True
    default_profile: str | None = None
    is_admin: bool = False


class UsersListUserResponse(UserBase):
    """Represents the response object for a single user in the list of users.

    This class inherits from the `UserBase` class and provides additional functionality specific to the response object.

    Attributes:
        Inherits all attributes from the `UserBase` class.

    """

    pass


class UsersListUserRequest(SQLModel):
    """Represents a request to list users.

    Attributes:
        name (str): The name of the user.
    """

    name: str


UsersListUsersResponse = Annotated[
    list[UsersListUserResponse],
    Field(..., description="A list of users"),
]


class UsersUpdateUserResponse(UserBase):
    """Represents the response for updating a user.

    This class inherits from the `UserBase` class and includes any additional
    attributes or methods specific to the response for updating a user.
    """

    pass


class UsersUpdateUserRequest(SQLModel):
    """Represents a request to update a user.

    Attributes:
        name (str): The name of the user.
        contact (EmailStr | None, optional): The contact email of the user. Defaults to None.
        new_email (EmailStr | None, optional): The new email of the user. Defaults to None.
        is_active (bool | None, optional): Indicates if the user is active. Defaults to None.
        default_profile (str | None, optional): The default profile of the user. Defaults to None.
    """

    name: str
    contact: EmailStr | None = None
    new_email: EmailStr | None = None
    is_active: bool | None = None
    default_profile: str | None = None
    is_admin: bool | None = None


class UsersDeleteUserResponse(SQLModel):
    """Represents the response for deleting a user.

    Attributes:
        ok (bool): Indicates whether the user deletion was successful or not.
    """

    ok: bool


class UsersDeleteUserRequest(SQLModel):
    """Represents a request to delete a user.

    Attributes:
        name (str): The name of the user to be deleted.
    """

    name: str


class UsersRotateTokenResponse(SQLModel):
    """Represents the response object for the token rotation operation in the Users API.

    Attributes:
        token (str): The new token generated after rotating the existing token.
    """

    token: str


class UsersRotateTokenRequest(SQLModel):
    """Represents a request to rotate the token for a user.

    Attributes:
        name (str): The name of the user.
    """

    name: str


class ProfilesCreateProfileRequest(SQLModel):
    """Represents a request to create a profile.

    Attributes:
        name (str): The name of the profile.
        checks (List[str], optional): A list of checks associated with the profile. Defaults to an empty list.
    """

    name: str
    checks: list[str] = []


class ProfilesCreateProfileResponse(ProfileBase):
    """Represents the response object for creating a profile.

    Attributes:
        checks (str): The list of names of checks associated with the profile.
    """

    checks: list[str] = []


class ProfilesListProfileRequest(SQLModel):
    """Represents a request to list profiles.

    Attributes:
        name (str): The name of the profile.
    """

    name: str


class ProfilesListProfileResponse(ProfileBase):
    """Represents a response object for listing profiles.

    Attributes:
        checks (list[ChecksListChecksResponse]): A list of checks associated with the profile.
    """

    checks: list["ChecksListChecksResponse"] = []


ProfilesListProfilesResponse = Annotated[
    list[ProfilesListProfileResponse],
    Field(..., description="A list of profiles"),
]


class ProfilesUpdateProfileRequest(SQLModel):
    """Represents a request to update a profile.

    Attributes:
        name (str): The name of the profile.
        checks (list[str], optional): A list of check names associated with the profile. Defaults to None.
    """

    name: str
    checks: list[str] | None = None


class ProfilesUpdateProfileResponse(ProfileBase):
    """Represents the response for updating a profile.

    This class inherits from the `ProfileBase` class and provides additional functionality
    specific to updating a profile.
    """

    pass


class ProfilesDeleteProfileRequest(SQLModel):
    """Represents a request to delete a profile.

    Attributes:
        name (str): The name of the profile to be deleted.
    """

    name: str


class ProfilesDeleteProfileResponse(SQLModel):
    """Represents the response for deleting a profile.

    This class inherits from the `SQLModel` class.
    """

    ok: bool


class ChecksCreateCheckResponse(CheckBase):
    """Represents the response object for creating a check.

    Attributes:
        profiles (str): The list of names of profiles associated with the check.
    """

    profiles: list[str] = []


class ChecksCreateCheckRequest(SQLModel):
    """Represents a request to create a check.

    Attributes:
        name (str): The name of the check.
        profiles (list[str], optional): A list of profiles associated with the check. Defaults to an empty list.
        upstream_services (list[str], optional): A list of upstream services for the check.
        config (dict[str, Union[str, int, float, bool]], optional): The configuration for the check. Defaults to an empty dictionary.
    """

    name: str
    profiles: list[str] = []
    upstream_services: list[str]
    supported_types: list[SafetyServicePayloadType] = []
    config: dict[str, Union[str, int, float, bool]] = {}


class ChecksListCheckResponse(CheckBase):
    """Represents the response for a single check in the list of checks.

    Attributes:
        profiles (list[str]): A list of profiles associated with the check.
    """

    profiles: list[str] = []


class ChecksListCheckRequest(SQLModel):
    """Represents a request to list checks for a specific check name."""

    name: str


ChecksListChecksResponse = Annotated[
    list[ChecksListCheckResponse],
    Field(..., description="A list of checks"),
]


class ChecksUpdateCheckResponse(CheckBase):
    """Represents the response for updating a check.

    This class inherits from the CheckBase class and provides additional functionality
    specific to updating a check.
    """

    pass


class ChecksUpdateCheckRequest(SQLModel):
    """Represents a request to update a check.

    Attributes:
        name (str): The name of the check.
        profiles (list[str], optional): A list of profiles associated with the check. Defaults to None.
        upstream_services (list[str], optional): A list of upstream services for the check. Defaults to None.
        config (dict[str, Union[str, int, float, bool]], optional): The configuration for the check. Defaults to None.
    """

    name: str
    profiles: list[str] | None = None
    upstream_services: list[str] | None = None
    config: dict[str, Union[str, int, float, bool]] | None = None


class ChecksDeleteCheckResponse(SQLModel):
    """Represents the response for deleting a check.

    Attributes:
        ok (bool): Indicates whether the check deletion was successful.
    """

    ok: bool


class ChecksDeleteCheckRequest(SQLModel):
    """Represents a request to delete a check.

    Attributes:
        name (str): The name of the check to be deleted.
    """

    name: str


class SafetyScanRequestWithText(SafetyServicePayloadText):
    """Represents a safety scan request with text.

    Attributes:
        kind (Literal["text"]): The type of payload, which is always "text".
    """

    kind: Literal["text"] = SafetyServicePayloadType.TEXT.value


class SafetyScanRequestWithImage(SafetyServicePayloadImage):
    """Represents a safety scan request with an image.

    Attributes:
        kind (Literal["image"]): The type of payload, which is always "image".
    """

    kind: Literal["image"] = SafetyServicePayloadType.IMAGE.value


SafetyScanRequest = Annotated[
    Union[SafetyScanRequestWithText, SafetyScanRequestWithImage],
    Field(..., discriminator="kind"),
]


class SafetyScanResponse(Verdict):
    """Represents the response from a safety scan.

    This class inherits from the `Verdict` class and provides additional functionality specific to safety scans.
    """

    pass

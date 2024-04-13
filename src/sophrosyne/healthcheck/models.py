"""Healthcheck models.

Models are based on the following expired IETF healthcheck draft:
https://datatracker.ietf.org/doc/html/draft-inadarei-api-health-check

Attributes:
    Status (Enum): The status of the health check.
    SubComponent (BaseModel): The subcomponent of the health check.
    Check (BaseModel): The check of the health check.
    HealthCheck (BaseModel): The health check.
"""

from datetime import datetime
from enum import Enum
from typing import Annotated, Any

from pydantic import (
    AwareDatetime,
    BaseModel,
    Field,
    ValidationInfo,
    field_validator,
    model_serializer,
)
from pydantic.config import ConfigDict


class Status(str, Enum):
    """Represents the status of a health check.

    Attributes:
        PASS (str): Indicates a successful health check.
        FAIL (str): Indicates a failed health check.
        WARN (str): Indicates a warning for a health check.
    """

    PASS = "pass"
    FAIL = "fail"
    WARN = "warn"


class SubComponent(BaseModel):
    """Represents a subcomponent of a health check.

    Attributes:
        component_id (str | None): The ID of the subcomponent.
        component_type (str | None): The type of the subcomponent.
        observed_value (str | int | float | list | dict | bool | None): The observed value of the subcomponent.
        observed_unit (str | None): The unit of measurement for the observed value.
        status (Status | None): The status of the subcomponent.
        affected_endpoints (list[str] | None): The endpoints affected by the subcomponent.
        time (AwareDatetime | None): The timestamp of when the subcomponent was observed.
        output (str | None): The output of the subcomponent.
        links (dict[str, str] | None): Additional links related to the subcomponent.
        additional_keys (dict[str, str] | None): Additional custom keys and values for the subcomponent.
    """

    component_id: str | None = Field(default=None, serialization_alias="componentId")
    component_type: str | None = Field(
        default=None, serialization_alias="componentType"
    )
    observed_value: str | int | float | list | dict | bool | None = Field(
        default=None, serialization_alias="observedValue"
    )
    observed_unit: str | None = Field(default=None, serialization_alias="observedUnit")
    status: Status | None = Field(default=None)
    affected_endpoints: list[str] | None = Field(
        default=None, serialization_alias="affectedEndpoints"
    )
    time: AwareDatetime | None = Field(default=None)
    output: str | None = Field(default=None)
    links: dict[str, str] | None = Field(default=None)

    additional_keys: dict[str, str] | None = Field(
        default=None, serialization_alias="additionalKeys"
    )

    @model_serializer
    def ser_model(self) -> dict[str, Any]:
        """Serializes the model object into a dictionary.

        Returns:
            dict[str, Any]: The serialized model as a dictionary.
        """
        out: dict[str, Any] = {}
        if self.component_id is not None:
            out["componentId"] = self.component_id
        if self.component_type is not None:
            out["componentType"] = self.component_type
        if self.observed_value is not None:
            out["observedValue"] = self.observed_value
        if self.observed_unit is not None:
            out["observedUnit"] = self.observed_unit
        if self.status is not None:
            out["status"] = self.status
        if self.affected_endpoints is not None and self.status != Status.PASS:
            out["affectedEndpoints"] = self.affected_endpoints
        if self.time is not None:
            out["time"] = self.time
        if self.output is not None and self.status != Status.PASS:
            out["output"] = self.output
        if self.links is not None:
            out["links"] = self.links

        if self.additional_keys is not None:
            for key, value in self.additional_keys.items():
                if key not in out:  # Do not overwrite existing keys
                    out[key] = value
        return out


class Check(BaseModel):
    """Represents a health check.

    Attributes:
        sub_components (dict[str, list[SubComponent]] | None): A dictionary mapping sub-component names to lists of SubComponent objects.
    """

    sub_components: dict[str, list[SubComponent]] | None = Field(default=None)

    @model_serializer
    def ser_model(self) -> dict[str, Any] | None:
        """Serialize the Check object to a dictionary.

        Returns:
            dict[str, Any] | None: The serialized Check object, or None if sub_components is None.
        """
        out: dict[str, Any] = {}
        if self.sub_components is None:
            return None
        for key, value in self.sub_components.items():
            out[key] = [v.ser_model() for v in value]

        return out


class HealthCheck(BaseModel):
    """Represents a health check object.

    Attributes:
        status (Status): The status of the health check.
        version (str | None): The version of the health check (default: None).
        release_ID (str | None): The release ID of the health check (default: None).
        notes (str | None): Additional notes for the health check (default: None).
        output (str | None): The output of the health check (default: None).
        checks (Check | None): The checks performed for the health check (default: None).
        links (dict[str, str] | None): Links related to the health check (default: None).
        service_id (str | None): The service ID of the health check (default: None).
        description (str | None): The description of the health check (default: None).
    """

    status: Status = Field()
    version: str | None = Field(default=None)
    release_ID: str | None = Field(default=None, serialization_alias="releaseId")
    notes: str | None = Field(default=None)
    output: str | None = Field(default=None)
    checks: Check | None = Field(default=None)
    links: dict[str, str] | None = Field(default=None)
    service_id: str | None = Field(default=None, serialization_alias="serviceId")
    description: str | None = Field(default=None)

    @model_serializer
    def ser_model(self) -> dict[str, Any]:
        """Serializes the model into a dictionary.

        Returns:
            dict[str, Any]: The serialized model.
        """
        out: dict[str, Any] = {"status": self.status}
        if self.version is not None:
            out["version"] = self.version
        if self.release_ID is not None:
            out["releaseId"] = self.release_ID
        if self.notes is not None:
            out["notes"] = self.notes
        if self.output is not None and self.status != Status.PASS:
            out["output"] = self.output
        if self.checks is not None:
            out["checks"] = self.checks.ser_model()
        if self.links is not None:
            out["links"] = self.links
        if self.service_id is not None:
            out["serviceId"] = self.service_id
        if self.description is not None:
            out["description"] = self.description
        return out

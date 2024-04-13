"""Safety module for performing safety checks on profiles."""

from enum import Enum
from typing import Iterable, Union

from fastapi import HTTPException
from pydantic import BaseModel, Field
from sqlmodel import select
from sqlmodel.ext.asyncio.session import AsyncSession

from sophrosyne.core.checks import Check
from sophrosyne.core.models import Profile, SafetyServicePayload, Verdict


class Safety:
    """Safety class for performing safety checks on profiles.

    Attributes:
        db_session (AsyncSession): The database session to use.
    """

    db_session: AsyncSession

    def __init__(self, db_session: AsyncSession) -> None:
        """Initialize the Safety class with a database session.

        Args:
            db_session (AsyncSession): The database session to use.
        """
        self.db_session = db_session

    async def predict(self, profile: str, data: SafetyServicePayload) -> Verdict:
        """Predict the safety verdict for a given profile and data.

        Args:
            profile (str): The name of the profile to check.
            data (SafetyServicePayload): The data to perform safety checks on.

        Returns:
            Verdict: The safety verdict.
        """
        result = await self.db_session.exec(
            select(Profile).where(Profile.name == profile)
        )
        db_profile = result.first()
        if not db_profile:
            raise HTTPException(status_code=404, detail="Profile not found")
        check_results: dict[str, bool] = {}
        for check in db_profile.checks:
            check_results[check.name] = Check.model_validate(check.model_dump()).run(
                data
            )

        # Bug / Point of Contention
        # If there are no checks associated with the profile, the verdict will
        # always be True. This may not be the desired behavior - if there are no
        # checks being performed, how can we assume it is safe? May be better to
        # return a False verdict in this case.

        return Verdict(verdict=_all(check_results.values()), checks=check_results)


def _all(iterable: Iterable[object]) -> bool:
    """Custom implementation of `all`.

    Differs from the built-in `all` in that it returns False if the iterable is
    empty.

    Args:
        iterable (Iterable[object]): The iterable to check.

    Returns:
        bool: True if all elements are truthy, False otherwise.
    """
    count_t: int = 0
    count_f: int = 0
    for item in iterable:
        if bool(item):
            count_t += 1
        else:
            count_f += 1
    return count_t > count_f

"""This module is responsible for creating the database and tables, and also for creating the root user."""

from alembic import command, config
from sqlalchemy.ext.asyncio import async_sessionmaker, create_async_engine
from sqlmodel import SQLModel, select
from sqlmodel.ext.asyncio.session import AsyncSession

from sophrosyne.core.config import get_settings
from sophrosyne.core.logging import get_logger
from sophrosyne.core.models import Profile, User
from sophrosyne.core.security import new_token, sign

engine = create_async_engine(
    get_settings().database.dsn,
    echo=get_settings().development.sqlalchemy_echo,
    future=True,
)

log = get_logger()


def alembic_config() -> config.Config:
    """Creates an returns a configuration for alembic."""
    cfg = config.Config()
    cfg.set_main_option("script_location", "sophrosyne:migrations")
    cfg.set_main_option("sqlalchemy.url", get_settings().database.dsn)

    return cfg


async def create_db_and_tables():
    """Create the database and tables."""
    cfg = alembic_config()

    def stamp(connection):
        cfg.attributes["connection"] = connection
        command.stamp(cfg, "head")

    async with engine.begin() as conn:
        await conn.run_sync(SQLModel.metadata.create_all)
        await conn.run_sync(stamp)


async def create_default_profile():
    """Create the default profile if it does not exist."""
    async_session = async_sessionmaker(
        engine, class_=AsyncSession, expire_on_commit=False
    )
    async with async_session() as session:
        result = await session.exec(
            select(Profile).where(Profile.name == get_settings().default_profile)
        )
        p = result.first()
        if p:
            return ""

        profile = Profile(name=get_settings().default_profile)
        session.add(profile)
        await session.commit()

        return profile


async def create_root_user() -> str:
    """Create the root user if it does not exist."""
    async_session = async_sessionmaker(
        engine, class_=AsyncSession, expire_on_commit=False
    )
    async with async_session() as session:
        result = await session.exec(
            select(User).where(User.contact == get_settings().root_contact)
        )
        u = result.first()
        if u:
            return ""

        token = new_token()
        if get_settings().development.static_root_token != "":
            token = get_settings().development.static_root_token
            log.warn("static root token in use")
        user = User(
            name="root",
            contact=get_settings().root_contact,
            signed_token=sign(token),
            is_active=True,
            is_admin=True,
        )
        session.add(user)
        await session.commit()

        return token


async def upgrade(revision: str):
    """Run database upgrade migration using alembic.

    Args:
       revision (str): The ID of the revision to upgrade the database do.
    """
    cfg = alembic_config()

    def _upgrade(revision: str):
        def execute(connection):
            cfg.attributes["connection"] = connection
            command.upgrade(cfg, revision)

        return execute

    async with engine.begin() as conn:
        await conn.run_sync(_upgrade(revision))


async def downgrade(revision: str):
    """Run database downgrade migration using alembic.

    Args:
       revision (str): The ID of the revision to downgrade the database do.
    """
    cfg = alembic_config()

    def _downgrade(revision: str):
        def execute(connection):
            cfg.attributes["connection"] = connection
            command.downgrade(cfg, revision)

        return execute

    async with engine.begin() as conn:
        await conn.run_sync(_downgrade(revision))


async def history(verbose: bool):
    """Show the database migration history.

    Args:
       verbose (bool): Be verbose in the output.
    """
    cfg = alembic_config()

    def show(connection):
        cfg.attributes["connection"] = connection
        command.history(cfg, verbose=verbose, indicate_current=True)

    async with engine.begin() as conn:
        await conn.run_sync(show)


async def current(verbose: bool):
    """Show the current database migration.

    Args:
       verbose (bool): Be verbose in the output.
    """
    cfg = alembic_config()

    def show(connection):
        cfg.attributes["connection"] = connection
        command.current(cfg, verbose=verbose)

    async with engine.begin() as conn:
        await conn.run_sync(show)

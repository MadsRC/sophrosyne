"""Configuration module for the SOPH API.

This module contains the configuration settings for the SOPH API. The settings
are loaded from environment variables, a YAML configuration file, and secrets
files. The settings are organized into classes that represent different parts
of the configuration. The settings are loaded using the `pydantic-settings`
library, which provides a way to load settings from multiple sources and
validate the settings.

The settings are loaded in the following order of priority:
1. Secrets files
2. Environment variables
3. YAML configuration file

Dynamic module behaviour:
    Because of a limitation in the `pydantic-settings` library, namely that it
    does not support dynamically setting model configuration (e.g., `secrets_dir`
    and `yaml_file`), the default values for these settings are set using
    environment variables. This allows the settings to be configured using
    dynamically set environment variables, although this is not ideal.

    The behaviour of this module that is considered dynamic are:

    - The path to the configuration file.
      The value used by this module is available as the `config_file` attribute,
      and the environment variable used to set this value is
      `SOPH__CONFIG_YAML_FILE`. The default value is `config.yaml`.
    - The path to the directory containing secrets files.
      The value used by this module is available as the `secrets_dir` attribute,
      and the environment variable used to set this value is
      `SOPH__CONFIG_SECRETS_DIR`. The default value is `/run/secrets` on Linux
      and an empty string on macOS. The empty string is used on macOS purely for
      testing purposes, as the `/run` directory is not available on macOS.
    - Whether to create the secrets directory if it does not exist.
      The value used by this module is available as the `create_secrets_dir`
      attribute, and the environment variable used to set this value is
      `SOPH__CONFIG_CREATE_SECRETS_DIR`. The default value is `false`.

    Environment variables used to dynamically alter the behaviour of this module
    is carefully chosen to avoid conflicts with the settings themselves. The
    prefix `SOPH__CONFIG_` is used to avoid conflicts.

    The module behaviour controlled by these environment variables is run once
    when the module is imported. The settings are then loaded using the
    configured settings sources.

Configuration via secrets files:
    The secrets files are loaded from the directory specified by the
    `secrets_dir` attribute. The files in these directories will be read and
    their literal contents will be used as the values for the settings. The
    secrets files are expected to be named after the settings they are providing
    secrets for, taking into account the environment prefix.

    This also means that the name of the secrets file for a setting is identical
    to the environment variable used to set the value of the setting. For
    example, the secrets file for the `password` setting in the `Database` class
    is expected to be named `SOPH_DATABASE__PASSWORD`.

    Configuration applied via will override all other sources of configuration.

Configuration via YAML configuration file:
    The YAML configuration file is loaded from the path specified by the
    `config_file` attribute. The settings in the YAML file are expected to be
    organized in the same structure as the settings classes in this module. The
    settings in the YAML file will be used to override the default values of the
    settings.

Configuration via environment variables:
    The settings can be configured using environment variables. The environment
    variables are expected to be named after the settings they are configuring,
    taking into account the environment prefix. The environment variables are
    expected to be in uppercase and use underscores to separate words. For
    example, the environment variable for the `password` setting in the
    `Database` class is expected to be named `SOPH_DATABASE__PASSWORD`.

    The environment variables will be used to override the default values of the
    settings as well as the values loaded from the YAML configuration file, but
    will be overridden by values loaded from secrets files.

    When providing lists via environment variables, the list should be a JSON
    array. For example, to provide a list of backend CORS origins, the
    environment variable should be named `SOPH_BACKEND_CORS_ORIGINS` and the
    value could be `'["http://localhost:3000", "http://localhost:3001"]'`.

Users of this module is encouraged to use the `get_settings` function to get the
settings object, as this function uses a LRU cache to ensure that the settings
are only loaded once.

Example:
    To get the settings object, use the `get_settings` function:
        settings = get_settings()

    The settings object can then be used to access the settings values:
        print(settings.database.host)


Attributes:
    config_file (str): The path to the configuration file.
    secrets_dir (str): The path to the directory containing secrets files.
    create_secrets_dir (bool): Whether to create the secrets directory if it
    does not exist.
"""

import base64
import logging
import os
import sys
from functools import lru_cache
from typing import Annotated, List, Literal, Tuple, Type

from pydantic import (
    AnyHttpUrl,
    Base64Encoder,
    EmailStr,
    EncodedBytes,
    Field,
    computed_field,
)
from pydantic_settings import (
    BaseSettings,
    PydanticBaseSettingsSource,
    SettingsConfigDict,
    YamlConfigSettingsSource,
)

config_file = os.environ.get("SOPH__CONFIG_YAML_FILE", "config.yaml")
secrets_dir = os.environ.get(
    "SOPH__CONFIG_SECRETS_DIR", ("" if sys.platform == "darwin" else "/run/secrets")
)
create_secrets_dir = (
    os.environ.get("SOPH__CONFIG_CREATE_SECRETS_DIR", "false").lower() == "true"
)

if create_secrets_dir and secrets_dir == "":
    raise ValueError(
        "Cannot create secrets dir when secrets dir is empty. Configure SOPH__CONFIG_SECRETS_DIR"
    )

if create_secrets_dir:
    os.makedirs(secrets_dir, exist_ok=True)


class Base64EncoderSansNewline(Base64Encoder):
    """Encode Base64 without adding a trailing newline.

    The default Base64Bytes encoder in PydanticV2 appends a trailing newline
    when encoding. See https://github.com/pydantic/pydantic/issues/9072
    """

    @classmethod
    def encode(cls, value: bytes) -> bytes:  # noqa: D102
        return base64.b64encode(value)


Base64Bytes = Annotated[bytes, EncodedBytes(encoder=Base64EncoderSansNewline)]


class Logging(BaseSettings):
    """Configuration class for the logging settings.

    Attributes of this class can be overriden by environment variables, a YAML
    configuration file, and secrets files. The environment variables and secret
    files must be named after the attributes they are setting, with the
    environment prefix and nested delimiter taken into account.

    The environment prefix is `SOPH_DATABASE__`.

    Attributes:
        level (Literal["info", "debug"]): The log level to use. Defaults to "info".
        event_field (str): The name of the field to use for the main part of the log. Defaults to "event".
        format (Literal["development", "production"]): The format of the logs. Defaults to "production".
    """

    level: Literal["info", "debug"] = "info"
    event_field: str = "event"
    format: Literal["development", "production"] = "production"

    @computed_field
    def level_as_int(self) -> int:
        """Provides the log_level as an integer."""
        if self.level.lower() == "debug":
            return logging.DEBUG
        else:
            return logging.INFO

    model_config = SettingsConfigDict(
        secrets_dir=secrets_dir, env_prefix="SOPH_logging__"
    )


class Database(BaseSettings):
    """Configuration class for the database settings.

    Attributes of this class can be overriden by environment variables, a YAML
    configuration file, and secrets files. The environment variables and secret
    files must be named after the attributes they are setting, with the
    environment prefix and nested delimiter taken into account.

    The environment prefix is `SOPH_DATABASE__`.

    Attributes:
        host (str): The host of the database.
        port (int): The port of the database.
        database (str): The name of the database.
        password (str): The password for the database.
        user (str): The user for the database.
    """

    host: str = "localhost"
    port: int = 5432
    database: str = "postgres"
    password: str = "postgres"
    user: str = "postgres"

    @computed_field
    def dsn(self) -> str:
        """Returns the Data Source Name (DSN) for connecting to the PostgreSQL database.

        The DSN is constructed using the user, password, host, port, and
        database attributes of the Config object.

        Returns:
            str: The Data Source Name (DSN) for connecting to the PostgreSQL
            database.
        """
        return f"postgresql+asyncpg://{self.user}:{self.password}@{self.host}:{self.port}/{self.database}"

    model_config = SettingsConfigDict(
        secrets_dir=secrets_dir, env_prefix="SOPH_database__"
    )


class Development(BaseSettings):
    """Configuration class for development environment.

    These settings should never be changed in production, as they are meant for
    development purposes only.

    Attributes of this class can be overriden by environment variables, a YAML
    configuration file, and secrets files. The environment variables and secret
    files must be named after the attributes they are setting, with the
    environment prefix and nested delimiter taken into account.

    The environment prefix is `SOPH_DEVELOPMENT__`.

    Attributes:
        static_root_token (str): Override the random token generated at first statup.
        sqlalchemy_echo (bool): Instruct SQLAlchemy to log SQL commands.
    """

    static_root_token: str = ""
    sqlalchemy_echo: bool = False

    model_config = SettingsConfigDict(
        secrets_dir=secrets_dir, env_prefix="SOPH_development__"
    )


_default_key = b"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="


class Security(BaseSettings):
    """Configuration class for security settings.

    Attributes of this class can be overriden by environment variables, a YAML
    configuration file, and secrets files. The environment variables and secret
    files must be named after the attributes they are setting, with the
    environment prefix and nested delimiter taken into account.

    The environment prefix is `SOPH_SECURITY__`.

    Attributes:
        token_length (int): The number of bytes to generate for tokens used by the application.
        site_key (bytes): The site key used for protection tokens at rest. Value must be provided as a base64 encoded string. Minimum length is 32 bytes, and maximum length is 64 bytes.
        salt (bytes): The salt used for protection tokens at rest. Value must be provided as a base64 encoded string. Minimum length is 32 bytes.
        certificate_path (str): The path to the certificate file. If empty, a self-signed certificate will be generated.
        key_path (str): The path to the key file. Used for TLS.
        key_password (str): The password for the key file.
        outgoing_tls_verify (bool): Whether to verify outgoing TLS connections.
        outgoing_tls_ca_path (str): The path to the CA certificate file for outgoing TLS connections.
    """

    token_length: int = 128
    site_key: Annotated[Base64Bytes, Field(min_length=32, max_length=64)] = _default_key
    salt: Annotated[Base64Bytes, Field(min_length=32)] = _default_key
    certificate_path: str | None = None
    key_path: str | None = None
    key_password: str | None = None
    outgoing_tls_verify: bool = True
    outgoing_tls_ca_path: str | None = None

    def assert_non_default_cryptographic_material(self) -> None:
        """Asserts that important cryptographic materials do not have a default value.

        This function must be called as soon as theres a slight possibility that
        the cryptographic key material provided by this class is needed.

        Raises:
           ValueError: If site_key or salt has the default value, a ValueError
           will be raised.
        """
        if self.site_key == b"\x00" * 32:
            raise ValueError("security.site_key must be set")
        if self.salt == b"\x00" * 32:
            raise ValueError("security.salt must be set")

    model_config = SettingsConfigDict(
        secrets_dir=secrets_dir, env_prefix="SOPH_security__"
    )


class Server(BaseSettings):
    """Configuration class for server settings.

    Attributes of this class can be overriden by environment variables, a YAML
    configuration file, and secrets files. The environment variables and secret
    files must be named after the attributes they are setting, with the
    environment prefix and nested delimiter taken into account.

    The environment prefix is `SOPH_SERVER__` and the nested delimiter is `__`.

    Attributes:
        port (int): The port to run the server on.
        listen_host (str): The host to listen on.
    """

    port: int = 8000
    listen_host: str = "0.0.0.0"

    model_config = SettingsConfigDict(
        secrets_dir=secrets_dir, env_prefix="SOPH_server__"
    )


class Settings(BaseSettings):
    """Represents the settings for the application.

    Attributes of this class can be overriden by environment variables, a YAML
    configuration file, and secrets files. The environment variables and secret
    files must be named after the attributes they are setting, with the
    environment prefix and nested delimiter taken into account.

    The environment prefix is `SOPH_`, and the nested delimiter is `__`.

    Attributes:
        api_v1_str (str): The API version string.
        backend_cors_origins (List[AnyHttpUrl]): The list of backend CORS
        origins.
        root_contact (EmailStr): The root contact email address.
        hostnames (List[str]): The list of hostnames that the server should respond to. If generating a certificate, these values are used as the Common Name (CN) and Subject Alternate Name (SAN) in the certificate.
        checks (Checks): The checks configuration.
        database (Database): The database configuration.
        development (Development): The development configuration.

    Methods:
        settings_customise_sources: Customize the sources for loading settings.
    """

    api_v1_str: str = "/v1"
    backend_cors_origins: List[AnyHttpUrl] = []
    default_profile: str = "default"

    root_contact: EmailStr = "replaceme@withareal.email"  # type: ignore NOSONAR
    hostnames: List[str] = ["localhost"]
    database: Database = Database()
    security: Security = Security()
    server: Server = Server()
    development: Development = Development()
    logging: Logging = Logging()

    model_config = SettingsConfigDict(
        yaml_file=config_file,
        secrets_dir=secrets_dir,
        env_prefix="SOPH_",
        env_nested_delimiter="__",
    )

    @classmethod
    def settings_customise_sources(
        cls,
        settings_cls: Type[BaseSettings],
        init_settings: PydanticBaseSettingsSource,
        env_settings: PydanticBaseSettingsSource,
        dotenv_settings: PydanticBaseSettingsSource,
        file_secret_settings: PydanticBaseSettingsSource,
    ) -> Tuple[PydanticBaseSettingsSource, ...]:
        """Customize the sources for loading settings.

        Creating this class method, which pydantic will call behind the scenes,
        allows us to customize the sources for loading settings. The return
        value of this method is a tuple of settings sources that will be used
        to load the settings. The order of the sources in the tuple defines
        the priority of the sources. The first source in the tuple has the
        highest priority, and the last source in the tuple has the lowest
        priority.

        Args:
            settings_cls (Type[BaseSettings]): The settings class.
            init_settings (PydanticBaseSettingsSource): The initial settings source.
            env_settings (PydanticBaseSettingsSource): The environment settings source.
            dotenv_settings (PydanticBaseSettingsSource): The dotenv settings source.
            file_secret_settings (PydanticBaseSettingsSource): The file secret settings source.

        Returns:
            Tuple[PydanticBaseSettingsSource, ...]: A tuple of customized settings sources.
        """
        return (
            init_settings,
            file_secret_settings,
            env_settings,
            YamlConfigSettingsSource(settings_cls),
        )


@lru_cache
def get_settings():
    """Retrieves the settings object.

    This function is backed by an LRU cache to ensure that the settings are
    only loaded once, and that the same settings object is returned on
    subsequent calls.

    The object returned by this function is not safe to modify, as it is shared
    between all users of this function.

    Returns:
        Settings: The settings object.
    """
    return Settings()

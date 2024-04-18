"""Internal things."""

#
# Do NOT import any modules from sophrosyne before making sure you've read the
# docstring of the _necessary_evil function.
#


def necessary_evil(path: str):
    """Run some initial setup.

    This code, as the name implies, is a necessary evil to make up for a
    missing feature, or perhaps my own personal shortcomings, of Pydantic. It
    does not seem possible to dynamically specify external configuration files
    via `model_config` in Pydantic, forcing us to have the value of the
    `yaml_file` argument be a variable that is set at module import time. This
    unfortunately creates the side effect that the location of the yaml file
    must be known the config module is imported.

    The way this is handled is to have this function take care of setting the
    necessary environment variables to configure the config module before
    importing it.

    It is imperative that this function is run before any other modules from
    sophrosyne is imported. This is because many other modules import the
    config module, and if that happens before this function is run, everything
    breaks.

    Additionally, because this function is run early and by pretty much all
    commands, it is also used to centralize other things such as initialization
    of logging.
    """
    import os

    if "SOPH__CONFIG_YAML_FILE" not in os.environ:
        os.environ["SOPH__CONFIG_YAML_FILE"] = path
    import sophrosyne.core.config  # noqa: F401
    from sophrosyne.core.config import get_settings
    from sophrosyne.core.logging import initialize_logging

    initialize_logging(
        log_level=get_settings().logging.level_as_int,
        format=get_settings().logging.format,
        event_field=get_settings().logging.event_field,
    )

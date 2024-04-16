"""The checks module contains the logic to run checks on the data supplied to the SOPH API service.

Checks may either be completely implemented in this module
or may call out to external services to perform the check.
"""

from random import choice

import grpc

from sophrosyne.core.models import (
    CheckBase,
    SafetyServicePayload,
    SafetyServicePayloadImage,
    SafetyServicePayloadText,
    SafetyServicePayloadType,
)
from sophrosyne.grpc.checks import checks_pb2, checks_pb2_grpc


class Check(CheckBase):
    """Base class for all checks.

    Attributes:
        name (str): The name of the check.
        description (str): A description of the check.
    """

    def __init__(self, name: str, description: str):
        """Initializes the Check class.

        Args:
            name (str): The name of the check.
            description (str): A description of the check.
        """
        self.name = name
        self.description = description

    def handle_dummy(self) -> bool:
        """Handle the dummy check.

        Will return with the result specified in the config. If no result is
        specified, will return False.

        Returns:
            bool: The result of the dummy check.
        """
        if self.config is None:
            return False
        if "result" not in self.config:
            return False

        if isinstance(self.config["result"], bool):
            return self.config["result"]
        if isinstance(self.config["result"], str):
            if self.config["result"].lower() == "true":
                return True
            if self.config["result"].lower() == "false":
                return False
        if isinstance(self.config["result"], int):
            return bool(self.config["result"])
        if isinstance(self.config["result"], float):
            return bool(self.config["result"])

        return False

    def type_is_supported(self, data: SafetyServicePayload) -> bool:
        """Determines if the check supports the given data type.

        Args:
            data (SafetyServicePayload): The data to check.

        Returns:
            bool: True if the check supports the data type, False otherwise.
        """
        if isinstance(data, SafetyServicePayloadText):
            return SafetyServicePayloadType.TEXT in self.supported_types
        if isinstance(data, SafetyServicePayloadImage):
            return SafetyServicePayloadType.IMAGE in self.supported_types
        return False

    def run(self, data: SafetyServicePayload) -> bool:
        """Run the check on the data.

        Args:
            data (SafetyServicePayload): The data to run the check on.

        Raises:
            NotImplementedError: If the check is not implemented for the given data type.

        Returns:
            bool: True if the data passes the check, False otherwise.
        """
        if not self.type_is_supported(data):
            raise NotImplementedError("Check not implemented for data type")
        if isinstance(data, SafetyServicePayloadText):
            rpc_payload = checks_pb2.CheckRequest(text=data.text)
        elif isinstance(data, SafetyServicePayloadImage):
            rpc_payload = checks_pb2.CheckRequest(image=data.image)
        else:
            raise NotImplementedError("Check not implemented for data type")
        if self.name.startswith("local:dummy:"):
            return self.handle_dummy()
        channel = grpc.insecure_channel(choice(self.upstream_services))
        stub = checks_pb2_grpc.CheckServiceStub(channel)
        call = stub.Check(rpc_payload)
        return call.result

from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class CheckRequest(_message.Message):
    __slots__ = ("text", "image")
    TEXT_FIELD_NUMBER: _ClassVar[int]
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    text: str
    image: str
    def __init__(self, text: _Optional[str] = ..., image: _Optional[str] = ...) -> None: ...

class CheckResponse(_message.Message):
    __slots__ = ("result", "details")
    RESULT_FIELD_NUMBER: _ClassVar[int]
    DETAILS_FIELD_NUMBER: _ClassVar[int]
    result: bool
    details: str
    def __init__(self, result: bool = ..., details: _Optional[str] = ...) -> None: ...

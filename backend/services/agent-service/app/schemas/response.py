from __future__ import annotations
from typing import Any
from pydantic import BaseModel


class ApiResponse(BaseModel):
    success:    bool
    message:    str
    statusCode: int
    data:       Any = None

    @classmethod
    def ok(cls, data: Any = None, message: str = "") -> "ApiResponse":
        return cls(success=True, message=message, statusCode=200, data=data)

    @classmethod
    def created(cls, data: Any = None, message: str = "") -> "ApiResponse":
        return cls(success=True, message=message, statusCode=201, data=data)

    @classmethod
    def error(cls, status_code: int, message: str) -> "ApiResponse":
        return cls(success=False, message=message, statusCode=status_code, data=None)
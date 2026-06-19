from pydantic import BaseModel

class ServiceOption(BaseModel):
    id:    str
    label: str

class AppContext(BaseModel):
    current_path:       str
    service_id:         str | None = None
    standard_id:        str | None = None
    assessment_id:      str | None = None
    available_services: list[ServiceOption] = []

class QueryRequest(BaseModel):
    question:    str
    session_id:  str | None = None
    app_context: AppContext | None = None
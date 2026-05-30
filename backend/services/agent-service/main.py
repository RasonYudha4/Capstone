from fastapi.responses import JSONResponse
import uvicorn
from fastapi import FastAPI, Request
from app.api import router
from app.schemas.response import ApiResponse

app = FastAPI()
app.include_router(router)

@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception) -> JSONResponse:
    body = ApiResponse.error(status_code=500, message=str(exc))
    return JSONResponse(status_code=500, content=body.model_dump())

if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=8001, reload=False)
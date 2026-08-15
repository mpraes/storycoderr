from fastapi import FastAPI

from app.api.chat import router

app = FastAPI(title="fastapi-rag-demo")
app.include_router(router)

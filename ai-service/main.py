import os
from contextlib import asynccontextmanager

from dotenv import load_dotenv
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from services.analyzer import AnalyzerService
from services.db import Database

load_dotenv()


@asynccontextmanager
async def lifespan(app: FastAPI):
    await app.state.db.connect()
    yield
    await app.state.db.disconnect()


app = FastAPI(
    title="Portfolio Index AI Service",
    version="1.0.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["GET"],
    allow_headers=["*"],
)

# Initialize shared services
app.state.db       = Database(os.getenv("DB_URL", ""))
app.state.analyzer = AnalyzerService(app.state.db)


@app.get("/health")
async def health():
    return {"status": "ok", "service": "ai-service"}


@app.get("/analyze/{symbol}")
async def analyze(symbol: str):
    """
    Returns full technical + ML analysis for a given index symbol.
    Results are cached in Redis for 2 minutes.
    """
    result = await app.state.analyzer.analyze(symbol.upper())
    return result

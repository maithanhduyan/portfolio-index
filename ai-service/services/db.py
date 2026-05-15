import os
from typing import Any

import asyncpg


class Database:
    """Thin async wrapper around asyncpg for TimescaleDB queries."""

    def __init__(self, dsn: str) -> None:
        # Convert SQLAlchemy-style URL to plain asyncpg DSN if needed
        self._dsn = dsn.replace("postgresql+asyncpg://", "postgresql://")
        self._pool: asyncpg.Pool | None = None

    async def connect(self) -> None:
        self._pool = await asyncpg.create_pool(
            dsn=self._dsn,
            min_size=2,
            max_size=10,
            command_timeout=30,
        )

    async def disconnect(self) -> None:
        if self._pool:
            await self._pool.close()

    async def fetch(self, query: str, *args: Any) -> list[dict]:
        assert self._pool, "Database not connected"
        rows = await self._pool.fetch(query, *args)
        return [dict(r) for r in rows]

    async def fetchrow(self, query: str, *args: Any) -> dict | None:
        assert self._pool, "Database not connected"
        row = await self._pool.fetchrow(query, *args)
        return dict(row) if row else None

from __future__ import annotations

from datetime import datetime, timezone
from typing import Any

import numpy as np
import pandas as pd
import ta as talib

from services.db import Database
from models.forecaster import StatsForecaster


class AnalyzerService:
    """
    Performs full technical analysis + ML forecasting for a market index.

    Indicators computed:
      - RSI (14), MACD (12/26/9)
      - Bollinger Bands (20, 2σ)
      - EMA 50 / EMA 200
      - ATR (14)
      - Support / Resistance (pivot-based)

    ML Forecast:
      - Prophet 7-day close price forecast
    """

    def __init__(self, db: Database) -> None:
        self._db = db
        self._forecaster = StatsForecaster()

    async def analyze(self, symbol: str) -> dict[str, Any]:
        rows = await self._db.fetch(
            """
            SELECT time, open, high, low, close, volume
            FROM candles
            WHERE symbol = $1
              AND time >= NOW() - INTERVAL '90 days'
            ORDER BY time ASC
            """,
            symbol,
        )

        if len(rows) < 30:
            return self._empty_response(symbol, "Insufficient data")

        df = pd.DataFrame(rows)
        df["time"]   = pd.to_datetime(df["time"])
        df = df.set_index("time").sort_index()
        df = df.astype({"open": float, "high": float, "low": float, "close": float, "volume": float})

        # ─── Technical Indicators (using `ta` library) ───────────
        close_s  = df["close"]
        high_s   = df["high"]
        low_s    = df["low"]
        volume_s = df["volume"]

        rsi_s   = talib.momentum.RSIIndicator(close=close_s, window=14).rsi()
        macd_i  = talib.trend.MACD(close=close_s, window_slow=26, window_fast=12, window_sign=9)
        bb_i    = talib.volatility.BollingerBands(close=close_s, window=20, window_dev=2)
        ema50_s = talib.trend.EMAIndicator(close=close_s, window=50).ema_indicator()
        ema200_s= talib.trend.EMAIndicator(close=close_s, window=200).ema_indicator()

        close     = float(close_s.iloc[-1])
        rsi       = self._safe_float(rsi_s.iloc[-1])
        macd_val  = self._safe_float(macd_i.macd().iloc[-1])
        macd_sig  = self._safe_float(macd_i.macd_signal().iloc[-1])
        macd_hist = self._safe_float(macd_i.macd_diff().iloc[-1])
        bb_upper  = self._safe_float(bb_i.bollinger_hband().iloc[-1])
        bb_lower  = self._safe_float(bb_i.bollinger_lband().iloc[-1])
        ema50     = self._safe_float(ema50_s.iloc[-1])
        ema200    = self._safe_float(ema200_s.iloc[-1])

        # ─── Support / Resistance (simple pivot) ─────────────────
        support, resistance = self._pivot_levels(df)

        # ─── Signal ──────────────────────────────────────────────
        signal, confidence = self._compute_signal(
            rsi=rsi,
            macd_hist=macd_hist,
            close=close,
            ema50=ema50,
            ema200=ema200,
            bb_upper=bb_upper,
            bb_lower=bb_lower,
        )

        # ─── Trend ───────────────────────────────────────────────
        trend = self._compute_trend(close, ema50, ema200)

        # ─── ML Forecast (7 days) ─────────────────────────────────
        forecast_7d = self._forecaster.forecast(df["close"], days=7)

        return {
            "symbol":       symbol,
            "signal":       signal,
            "confidence":   confidence,
            "trend":        trend,
            "support":      round(support, 4),
            "resistance":   round(resistance, 4),
            "rsi":          round(rsi, 2),
            "macd": {
                "value":     round(macd_val, 6),
                "signal":    round(macd_sig, 6),
                "histogram": round(macd_hist, 6),
            },
            "bollinger": {
                "upper": round(bb_upper, 4),
                "lower": round(bb_lower, 4),
            },
            "ema":    {"ema50": round(ema50, 4), "ema200": round(ema200, 4)},
            "forecast_7d": [round(p, 4) for p in forecast_7d],
            "summary":      self._build_summary(signal, trend, rsi, macd_hist, forecast_7d, close),
            "generated_at": datetime.now(timezone.utc).isoformat(),
        }

    # ─── Helpers ─────────────────────────────────────────────────

    @staticmethod
    def _safe_float(val: Any) -> float:
        try:
            f = float(val)
            return f if not np.isnan(f) else 0.0
        except (TypeError, ValueError):
            return 0.0

    @staticmethod
    def _pivot_levels(df: pd.DataFrame) -> tuple[float, float]:
        """Simple pivot-point support/resistance from last 20 candles."""
        recent = df.tail(20)
        support    = float(recent["low"].min())
        resistance = float(recent["high"].max())
        return support, resistance

    @staticmethod
    def _compute_signal(
        rsi: float, macd_hist: float, close: float,
        ema50: float, ema200: float, bb_upper: float, bb_lower: float,
    ) -> tuple[str, int]:
        score = 0

        if rsi < 30:  score += 2
        elif rsi > 70: score -= 2

        if macd_hist > 0: score += 1
        elif macd_hist < 0: score -= 1

        if ema50 > 0 and ema200 > 0:
            if close > ema50 > ema200: score += 2
            elif close < ema50 < ema200: score -= 2

        if bb_lower > 0 and close <= bb_lower: score += 1
        if bb_upper > 0 and close >= bb_upper: score -= 1

        confidence = min(abs(score) * 15, 95)
        if score > 0:   return "BUY",     confidence
        if score < 0:   return "SELL",    confidence
        return "NEUTRAL", 50

    @staticmethod
    def _compute_trend(close: float, ema50: float, ema200: float) -> str:
        if ema50 > 0 and ema200 > 0:
            if close > ema50 > ema200: return "UPTREND"
            if close < ema50 < ema200: return "DOWNTREND"
        return "SIDEWAYS"

    @staticmethod
    def _build_summary(
        signal: str, trend: str, rsi: float,
        macd_hist: float, forecast_7d: list[float], close: float,
    ) -> str:
        direction = "tăng" if forecast_7d and forecast_7d[-1] > close else "giảm"
        rsi_desc  = "quá bán" if rsi < 30 else ("quá mua" if rsi > 70 else "trung tính")
        return (
            f"Xu hướng {trend.lower().replace('trend', ' trend')}, RSI {rsi:.1f} ({rsi_desc}), "
            f"MACD histogram {'dương' if macd_hist >= 0 else 'âm'}. "
            f"Tín hiệu: {signal}. Dự báo 7 ngày tới giá sẽ {direction}."
        )

    @staticmethod
    def _empty_response(symbol: str, reason: str) -> dict:
        return {
            "symbol": symbol,
            "signal": "NEUTRAL", "confidence": 0,
            "trend": "SIDEWAYS",
            "support": 0.0, "resistance": 0.0,
            "rsi": 0.0,
            "macd": {"value": 0.0, "signal": 0.0, "histogram": 0.0},
            "bollinger": {"upper": 0.0, "lower": 0.0},
            "ema": {"ema50": 0.0, "ema200": 0.0},
            "forecast_7d": [],
            "summary": reason,
            "generated_at": datetime.now(timezone.utc).isoformat(),
        }

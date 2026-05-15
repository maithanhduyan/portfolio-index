from __future__ import annotations

import warnings
import numpy as np
import pandas as pd
from statsmodels.tsa.holtwinters import ExponentialSmoothing


class StatsForecaster:
    """
    Uses Holt-Winters Exponential Smoothing to forecast N days of close prices.
    Pure Python — no C++ compilation needed, ~10x faster to install than Prophet.
    Falls back to linear extrapolation if data is insufficient.
    """

    def forecast(self, close_series: pd.Series, days: int = 7) -> list[float]:
        try:
            return self._holt_winters(close_series, days)
        except Exception:
            return self._linear_fallback(close_series, days)

    @staticmethod
    def _holt_winters(close_series: pd.Series, days: int) -> list[float]:
        series = close_series.dropna()
        if len(series) < 14:
            raise ValueError("Not enough data")

        with warnings.catch_warnings():
            warnings.simplefilter("ignore")
            model = ExponentialSmoothing(
                series,
                trend="add",
                seasonal=None,       # no fixed seasonality for financial data
                damped_trend=True,
                initialization_method="estimated",
            )
            fit = model.fit(optimized=True, remove_bias=True)

        return fit.forecast(days).tolist()

    @staticmethod
    def _linear_fallback(close_series: pd.Series, days: int) -> list[float]:
        y = close_series.dropna().values[-30:]
        if len(y) == 0:
            return [0.0] * days
        x = np.arange(len(y))
        coeffs = np.polyfit(x, y, 1)
        future_x = np.arange(len(y), len(y) + days)
        return np.polyval(coeffs, future_x).tolist()

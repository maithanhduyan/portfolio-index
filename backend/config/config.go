package config

import "os"

// Config holds all application configuration.
type Config struct {
	Port           string
	DBUrl          string
	RedisUrl       string
	AIServiceUrl   string
	AllowedOrigins string
	JWTSecret      string

	// External data source APIs
	BinanceWSUrl    string
	AlphaVantageKey string
	TCBSBaseUrl     string
	CoinGeckoURL    string
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		DBUrl:           getEnv("DB_URL", "postgres://postgres:secret@localhost:5432/portfolio?sslmode=disable"),
		RedisUrl:        getEnv("REDIS_URL", "redis://localhost:6379"),
		AIServiceUrl:    getEnv("AI_SERVICE_URL", "http://localhost:8000"),
		AllowedOrigins:  getEnv("ALLOWED_ORIGINS", "*"),
		JWTSecret:       getEnv("JWT_SECRET", "change-me-in-production-use-32chars"),
		BinanceWSUrl:    getEnv("BINANCE_WS_URL", "wss://stream.binance.com:9443/ws"),
		AlphaVantageKey: getEnv("ALPHA_VANTAGE_KEY", ""),
		TCBSBaseUrl:     getEnv("TCBS_BASE_URL", "https://apipubaws.tcbs.com.vn"),
		CoinGeckoURL:    getEnv("COINGECKO_URL", "https://api.coingecko.com/api/v3"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

# Portfolio Index

Trang web theo dõi chỉ số danh mục đầu tư theo thời gian thực — SPX500, VN30, VNINDEX, Bitcoin, Ethereum và hơn thế nữa.

---

## Kiến trúc

```
Binance WS  ──┐
TCBS API    ──┼──► Go Fetcher ──► TimescaleDB ──► Redis Pub/Sub ──► WebSocket ──► Browser
Alpha Vantage─┘                                                  └──► REST API
                                              AI Service (Python/Prophet)
```

## Tech Stack

| Layer           | Công nghệ                            | Lý do                                     |
|-----------------|--------------------------------------|-------------------------------------------|
| **Backend**     | Go 1.22 + Fiber v2                   | ~50MB RAM, 1M+ concurrent WS connections  |
| **AI/ML**       | Python 3.12 + FastAPI + Prophet      | Phân tích kỹ thuật + dự báo giá 7 ngày   |
| **Frontend**    | Next.js 14 + TailwindCSS + LW-Charts | TradingView-style candlestick charts      |
| **Time-series** | TimescaleDB (PostgreSQL hypertable)  | OHLCV query 10x nhanh hơn Postgres thuần |
| **Cache/PubSub**| Redis 7                              | <5ms latency, WebSocket broadcast         |
| **Proxy**       | Nginx                                | Rate limit, WS upgrade, static cache     |
| **Container**   | Docker Compose                       | One-command deploy                        |

## Nguồn dữ liệu

| Chỉ số          | Nguồn                        | Ghi chú               |
|-----------------|------------------------------|-----------------------|
| BTC, ETH        | Binance WebSocket (miễn phí) | Real-time tick data   |
| SPX, NDX        | Alpha Vantage API            | Free tier: 25 req/day |
| VNINDEX, VN30   | TCBS Public API              | Poll mỗi 10 giây      |

## Cài đặt & Chạy

### Yêu cầu
- Docker Desktop

### Bước 1 — Tạo file `.env`
```bash
cp .env.example .env
# Điền ALPHA_VANTAGE_KEY nếu muốn dữ liệu US stocks
```

### Bước 2 — Khởi động
```bash
docker compose up --build
```

Truy cập: **http://localhost**

### Development (không dùng Docker)
```bash
make backend-run      # Go API :8080
make frontend-dev     # Next.js :3000
make ai-dev           # Python AI :8000
```

## Cấu trúc project

```
portfolio-index/
├── backend/                  # Go + Fiber
│   ├── config/               # Cấu hình từ env
│   ├── handlers/             # REST handlers + WebSocket Hub
│   ├── models/               # Domain models
│   ├── repository/           # TimescaleDB queries
│   ├── services/             # Redis cache + Data fetcher
│   └── main.go
├── frontend/                 # Next.js 14
│   └── src/
│       ├── app/              # App Router pages
│       ├── components/       # IndexCard, PriceChart, MarketOverview
│       ├── lib/              # API client, WebSocket store (Zustand)
│       └── types/            # TypeScript interfaces
├── ai-service/               # Python + FastAPI
│   ├── models/               # Prophet forecaster
│   └── services/             # Technical analyzer (RSI, MACD, BB)
└── infrastructure/
    ├── db/init.sql           # TimescaleDB schema + hypertables
    └── nginx/nginx.conf      # Reverse proxy + rate limiting
```

## API Endpoints

| Method | Endpoint                              | Mô tả                       |
|--------|---------------------------------------|-----------------------------|
| GET    | `/api/v1/indices`                     | Tất cả chỉ số (cached 5s)   |
| GET    | `/api/v1/indices/:symbol`             | Chi tiết 1 chỉ số           |
| GET    | `/api/v1/indices/:symbol/candles`     | OHLCV data (có phân trang)  |
| GET    | `/api/v1/indices/:symbol/stats`       | Thống kê 7d/30d/1y          |
| GET    | `/api/v1/indices/:symbol/ai-analysis` | RSI, MACD, dự báo 7 ngày    |
| WS     | `/ws/prices`                          | Real-time price stream      |

## Scale-up cho 20 triệu users

- **Go Fiber**: non-blocking I/O, goroutine-per-connection, ~50MB RAM cho 10K WS clients
- **Redis Pub/Sub**: 1 publish → broadcast tới toàn bộ WS clients, không loop qua DB
- **TimescaleDB continuous aggregates**: OHLCV 1h/1d được pre-compute, query không scan raw data
- **Nginx**: rate limiting 60 req/min/IP, keepalive 1000 req/connection, gzip compression
- **Horizontal scaling**: chạy N backend container sau Nginx upstream, Redis làm shared state

.PHONY: dev build down logs clean

dev:
	docker compose up --build

build:
	docker compose build

down:
	docker compose down

logs:
	docker compose logs -f

clean:
	docker compose down -v --remove-orphans

# ─── Backend ──────────────────────────────────────────────────
backend-run:
	cd backend && go run main.go

backend-test:
	cd backend && go test ./...

backend-lint:
	cd backend && golangci-lint run

# ─── Frontend ─────────────────────────────────────────────────
frontend-dev:
	cd frontend && npm run dev

frontend-install:
	cd frontend && npm install

# ─── AI Service ───────────────────────────────────────────────
ai-dev:
	cd ai-service && uvicorn main:app --reload --port 8000

ai-install:
	cd ai-service && pip install -r requirements.txt

# ─── Database ─────────────────────────────────────────────────
db-migrate:
	docker compose exec timescaledb psql -U postgres -d portfolio -f /docker-entrypoint-initdb.d/01_init.sql

db-shell:
	docker compose exec timescaledb psql -U postgres -d portfolio

# Codyssey Makefile

include .env

.PHONY: help infra-up infra-down full-up full-down dev-go dev-py dev-frontend test-go test-py logs-go logs-py lint

help:
	@echo "Targets:"
	@echo "  infra-up       启动基础设施 (Hybrid)"
	@echo "  infra-down     关闭基础设施 (Hybrid)"
	@echo "  full-up        全容器启动 (包括应用)"
	@echo "  full-down      停止并移除全容器"
	@echo "  dev-go         本地运行 Go 后端"
	@echo "  dev-py         本地运行 Python AI 服务"
	@echo "  dev-frontend   本地运行前端"
	@echo "  test-go        运行 Go 测试"
	@echo "  test-py        运行 Python 测试"
	@echo "  logs-go        查看 Go 容器日志 (Full Docker 模式)"
	@echo "  logs-py        查看 Python 容器日志 (Full Docker 模式)"

infra-up:
	docker-compose -f infra/docker-compose.infra.yml up -d

infra-down:
	docker-compose -f infra/docker-compose.infra.yml down

full-up:
	cd infra && docker-compose up -d --build

full-down:
	cd infra && docker-compose down

dev-go:
	cd backend && GO_BACKEND_PORT=$${GO_BACKEND_PORT:-8080} go run .

dev-py:
	cd python && PY_BACKEND_PORT=$${PY_BACKEND_PORT:-8000} uvicorn main:app --reload --port $${PY_BACKEND_PORT}

dev-frontend:
	cd frontend && pnpm install && pnpm dev

test-go:
	cd backend && go test ./...

test-py:
	cd python && pytest -q

logs-go:
	cd infra && docker-compose logs -f go-backend

logs-py:
	cd infra && docker-compose logs -f py-backend

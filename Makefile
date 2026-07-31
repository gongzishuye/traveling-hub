COMPOSE = docker compose --env-file .env -f deploy/docker-compose.yml

.PHONY: up down logs test integration fmt

up:
	cp .env.example .env 2>/dev/null || true
	$(COMPOSE) up --build -d

down:
	$(COMPOSE) down -v

logs:
	$(COMPOSE) logs -f app

test:
	docker run --rm -e GOPROXY=https://goproxy.cn,direct -v "$(CURDIR):/app" -w /app golang:1.25 sh -lc 'export PATH=/usr/local/go/bin:$$PATH; go test ./...'

integration:
	$(COMPOSE) exec -T app go test -tags=integration ./tests/integration -count=1

fmt:
	docker run --rm -v "$(CURDIR):/app" -w /app golang:1.25 sh -lc 'export PATH=/usr/local/go/bin:$$PATH; gofmt -w $$(find . -name "*.go" -not -path "./.git/*")'

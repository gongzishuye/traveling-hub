COMPOSE = docker compose --env-file .env -f deploy/docker-compose.yml

.PHONY: up down logs test integration fmt verify-autonomous-traveler

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
	$(COMPOSE) up -d postgres redis
	$(COMPOSE) exec -T postgres sh -lc 'psql -U travelinghub -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname = '\''travelinghub_test'\''" | grep -q 1 || createdb -U travelinghub travelinghub_test'
	docker run --rm --network travelinghub_default -e GOPROXY=https://goproxy.cn,direct -e GOSUMDB=off -e TRAVELINGHUB_POSTGRES_DSN='postgres://travelinghub:travelinghub@postgres:5432/travelinghub_test?sslmode=disable' -e TRAVELINGHUB_REDIS_ADDR='redis:6379' -v travelinghub-go-mod:/go/pkg/mod -v travelinghub-go-build:/root/.cache/go-build -v "$(CURDIR):/app" -w /app golang:1.25 sh -lc 'export PATH=/usr/local/go/bin:$$PATH; go test -count=1 -tags=integration ./tests/integration && go test -count=1 -tags=integration ./internal/journey'

fmt:
	docker run --rm -v "$(CURDIR):/app" -w /app golang:1.25 sh -lc 'export PATH=/usr/local/go/bin:$$PATH; gofmt -w $$(find . -name "*.go" -not -path "./.git/*")'

# Runs every repository gate from the current working tree.  Frontend checks
# execute in a disposable Playwright container so no dependency or build
# directory is written back to the repository.
verify-autonomous-traveler: test integration
	test ! -e frontend/node_modules && test ! -e frontend/dist && test ! -e frontend/test-results && test ! -e frontend/.git
	docker run --rm -v "$(CURDIR)/frontend:/source:ro" mcr.microsoft.com/playwright:v1.62.0-noble bash -lc 'cp -a /source /work/frontend && cd /work/frontend && npm ci && npm test -- --run && npm run lint && npm run build && npm run test:e2e'

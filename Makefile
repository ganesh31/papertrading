.PHONY: up down logs migrate check dev

up:
	docker compose -f infra/docker-compose.yml up -d

down:
	docker compose -f infra/docker-compose.yml down

logs:
	docker compose -f infra/docker-compose.yml logs -f --tail=200

migrate:
	DATABASE_URL=$${DATABASE_URL:-postgres://papertrading:papertrading@localhost:5432/papertrading?sslmode=disable} ./infra/bin/dbmate.sh up

check:
	pnpm check
	go test ./...
	go build -o /dev/null ./cmd/pt
	(cd services/go/md && go test ./...)
	(cd services/go/matching && go test ./...)

dev:
	pnpm dev

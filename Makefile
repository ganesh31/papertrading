.PHONY: up down logs migrate proto check dev

BUF_VERSION := v1.50.0

proto:
	cd packages/protos && go run github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION) lint && \
		go run github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION) generate
	@git diff --quiet -- services/go/matching/pb || \
		(echo "protobuf codegen drift: run make proto and commit services/go/matching/pb" && git diff -- services/go/matching/pb && exit 1)

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
	$(MAKE) proto
	go test ./...
	go build -o /dev/null ./cmd/pt
	(cd services/go/md && go test ./...)
	(cd services/go/matching && go test ./...)

dev:
	pnpm dev

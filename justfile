set dotenv-load := false

# Run `just` from the repository root (same directory as this justfile).

instruments-sync:
	curl -fsSL 'https://margincalculator.angelbroking.com/OpenAPI_File/files/OpenAPIScripMaster.json' \
		-o infra/seed/scrip-master.json
	go run ./cmd/pt instruments sync --file infra/seed/scrip-master.json

data-fetch-minute symbols='RELIANCE,INFY,TCS,HDFCBANK,ICICIBANK,SBIN' days='7':
	go run ./cmd/pt data fetch minute --symbols '{{symbols}}' --days '{{days}}'

data-fetch-bhavcopy from='2026-04-10' to='2026-04-20':
	go run ./cmd/pt data fetch bhavcopy --from '{{from}}' --to '{{to}}'

data-refresh-all: data-fetch-minute data-fetch-bhavcopy

up:
  make up

down:
  make down

logs:
  make logs

migrate:
  make migrate

check:
  make check

dev:
  make dev

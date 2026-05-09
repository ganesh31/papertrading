set dotenv-load := false

# Run `just` from the repository root (same directory as this justfile).

instruments-sync:
	curl -fsSL 'https://margincalculator.angelbroking.com/OpenAPI_File/files/OpenAPIScripMaster.json' \
		-o infra/seed/scrip-master.json
	go run ./cmd/pt instruments sync --file infra/seed/scrip-master.json

# Positional args only (not symbols=...): first = CSV symbols, second = days.
# e.g.  just data-fetch-minute INFY,RELIANCE 7
data-fetch-minute symbols='RELIANCE,INFY,TCS,HDFCBANK,ICICIBANK,SBIN' days='7':
	go run ./cmd/pt data fetch minute --symbols '{{symbols}}' --days '{{days}}'

# Positional: first = --from (YYYY-MM-DD), second = --to (YYYY-MM-DD).
# e.g.  just data-fetch-bhavcopy 2026-04-10 2026-04-20
data-fetch-bhavcopy from='2026-04-10' to='2026-04-20':
	go run ./cmd/pt data fetch bhavcopy --from '{{from}}' --to '{{to}}'

data-refresh-all: data-fetch-minute data-fetch-bhavcopy

# Virtual-clock replay (requires md up with MD_ADAPTER=nse_replay).
replay date='2026-04-24' symbols='INFY,RELIANCE' speed='100':
	go run ./cmd/pt replay --date '{{date}}' --symbols '{{symbols}}' --speed '{{speed}}'

replay-stop:
	go run ./cmd/pt replay stop

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

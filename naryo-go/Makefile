MODULES    := core broadcaster-http broadcaster-kafka broadcaster-rabbitmq persistence-gorm server
BUILD_TAGS := test

GORM_DIR      := persistence-gorm
GORM_ATLAS    := $(GORM_DIR)/resources/atlas
NAME          ?= migration

.PHONY: all fmt fmt-check lint build test schema-diff-gorm schema-validate-gorm hooks-install \
	compose-up compose-down compose-logs compose-restart \
	$(addprefix fmt-,$(MODULES)) \
	$(addprefix lint-,$(MODULES)) \
	$(addprefix build-,$(MODULES)) \
	$(addprefix test-,$(MODULES))

all: fmt lint build test

# ------------------------------------------------------------------
# Format
# ------------------------------------------------------------------

fmt:
	gofmt -w .

fmt-check:
	@files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then \
		echo "The following files are not gofmt-formatted:"; \
		echo "$$files"; \
		echo "Run 'make fmt' to fix them."; \
		exit 1; \
	fi

$(addprefix fmt-,$(MODULES)): fmt-%:
	gofmt -w ./$*

# ------------------------------------------------------------------
# Lint
# ------------------------------------------------------------------

lint:
	golangci-lint run --build-tags $(BUILD_TAGS) $(foreach m,$(MODULES),./$(m)/...)

$(addprefix lint-,$(MODULES)): lint-%:
	golangci-lint run --build-tags $(BUILD_TAGS) ./$*/...

# ------------------------------------------------------------------
# Build
# ------------------------------------------------------------------

build:
	go build $(foreach m,$(MODULES),./$(m)/...)

$(addprefix build-,$(MODULES)): build-%:
	go build ./$*/...

# ------------------------------------------------------------------
# Test
# ------------------------------------------------------------------

test: $(addprefix test-,$(MODULES))

$(addprefix test-,$(MODULES)): test-%:
	go test -tags $(BUILD_TAGS) -count=1 ./$*/...

# ------------------------------------------------------------------
# Docker Compose (naryo-server + postgres + surfpool local Solana RPC)
# ------------------------------------------------------------------
compose-up:
	docker compose up -d --build

compose-down:
	docker compose --profile '*' down -v

compose-logs:
	docker compose logs -f

compose-restart: compose-down compose-up

# Brings up the full stack against the quickstart example config, including its dockerized
# receiver (quickstart-receiver, gated behind the "quickstart" profile so a plain compose-up/
# compose-restart never starts it) on a separate Postgres port so it doesn't collide with a plain
# `make compose-up`.
example-quickstart:
	CONFIG_PATH=examples/quickstart/application.yaml \
	POSTGRES_PORT=5433 \
	BROADCASTER_URL=http://quickstart-receiver:8081 \
	COMPOSE_PROFILES=quickstart \
	docker compose up -d --build

# ------------------------------------------------------------------
# Migrations (Atlas + GORM provider)
# ------------------------------------------------------------------
schema-diff-gorm:
	@command -v atlas >/dev/null 2>&1 || { echo "atlas CLI not found: https://atlasgo.io/getting-started"; exit 1; }
	@cd $(GORM_ATLAS) && atlas migrate diff $(NAME) --env gorm

schema-validate-gorm:
	@command -v atlas >/dev/null 2>&1 || { echo "atlas CLI not found: https://atlasgo.io/getting-started"; exit 1; }
	@cd $(GORM_ATLAS) && atlas migrate validate --env gorm

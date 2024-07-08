SRC         := $(shell find . -name '*.go')
QUERIES     := $(shell find . -name 'queries.sql')
QUERIES_GEN := internal/database/queries.sql.go internal/database/models.go internal/database/db.go

SQLC_VERSION ?= v1.26

.PHONY: build migrations queries

build: notes

notes: go.mod go.sum $(SRC) queries
	go build

migrations:
	mkdir -p ops/db/migrations
	cd ops/db && atlas schema fmt
	cd ops/db && atlas migrate diff --env local
	cd ops/db && atlas migrate lint --env local --git-base master

queries: $(QUERIES_GEN)
$(QUERIES_GEN) &: $(QUERIES) sqlc.yaml
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) generate

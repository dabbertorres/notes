QUERIES     := $(shell find . -name 'queries.sql')
QUERIES_GEN := $(QUERIES:%/queries.sql=%/db/queries.sql.go)

$(QUERIES_GEN): %/db/queries.sql.go: %/queries.sql sqlc.yaml
	sqlc generate

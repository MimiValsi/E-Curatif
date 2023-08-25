include .envrc

# ==================== #
# INTRODUCTION
# ==================== #
.PHONY: intro
intro:
	@echo "Welcome to E-Curatif!"
	@echo

# ==================== # 
# HELPERS
# ==================== # 

## help: print this message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# ==================== # 
# DEVELOPMENT
# ==================== # 

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	@psql $(ECURATIF_DB_DSN)

## run: run e-curatif/cmd app (Dev only)
.PHONY: run
# Only for test
run:
	@go run ./cmd/ecuratif/ -db-dsn=$(ECURATIF_DB_DSN)

# ==================== # 
# PRODUCTION
# ==================== # 

## build: build the program. Use it only for prod!
.PHONY: build
build:
	@go build -o launch ./cmd/e-curatif/ -db-dsn=$(ECURATIF_DB_DSN)

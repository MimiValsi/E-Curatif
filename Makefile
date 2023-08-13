FILE_PATH = $$GOPATH/bin/migrate

.PHONY: help
help:
	@echo "Welcome to E-Curatif!"
	@echo
	@echo "First time usage:"
	@echo "$$ make go-migrate"
	@echo "It will check if migrate file exists. If not, then it will be downloaded."
	@echo "It will be needed to trace PostgreSQL migrations."

# Check if go-migrate exists
.PHONY: go-migrate
go-migrate:
	@if [ -f $(FILE_PATH) ]; then \
		echo "migrate file exists"; \
		else \
		echo "nok"; \
		curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz ; \
		mv migrate.linux-amd64 $(FILE_PATH) ; \
		fi

.PHONY: build
build:
	go build ./cmd/app/
	go run ./cmd/migrator/main.go --storage=./storage/url_profile.db --migration-path=./migrations

.PHONY: run
run:
	go run ./cmd/sso/ --config ./config/local.yaml

.PHONY: migration
migration:
	go run ./cmd/migrator/main.go --storage=./storage/url_profile.db --migration-path=./migrations

.DEFAULT_GOAL := build
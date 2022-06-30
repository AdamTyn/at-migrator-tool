GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)
CONF_PROTO_FILES=$(shell find internal/conf -name *.proto)

.PHONY: init
# get default config.json
init:
	awk 'BEGIN { cmd="cp -ri ./config.json.example ./config.json"; print "n" |cmd; }'

.PHONY: config
# generate config proto
config:
	protoc --proto_path=./internal/conf \
 	       --go_out=paths=source_relative:./internal/conf \
	       $(CONF_PROTO_FILES)

.PHONY: run
# go run main.go
run:
	cd ./cmd && go run .

.PHONY: build
# go build main.go
build:
	cd ./cmd && go build -o ../bin/at-migrator-tool .

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z0-9_-]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
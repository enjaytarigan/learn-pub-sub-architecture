PWD := $(shell pwd)

.PHONY: run/server
  run/server:
	go build -o $(PWD)/bin/server $(PWD)/cmd/server && $(PWD)/bin/server

.PHONY: run/client
  run/client:
	go build -o $(PWD)/bin/client $(PWD)/cmd/client && $(PWD)/bin/client
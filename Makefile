default: build

build:
	go build

docs:
	go generate

test:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

fmt:
	find . -type f -name \*.go | xargs gofmt -w

.PHONY: default build docs test fmt

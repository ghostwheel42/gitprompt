.PHONY: all
all: test build

.PHONY: build
build:
	go build cmd/gitprompt/gitprompt.go

.PHONY: test
test:
	go test

.PHONY: clean
clean:
	rm -f gitprompt


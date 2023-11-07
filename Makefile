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
	rm -rf gitprompt release

.PHONY: release
release:
	rm -rf release/*
	mkdir -p release/files
	cp LICENSE README.md release/files/
	r="$$(git describe --tags)"; \
	for oa in darwin_arm64 darwin_amd64 linux_amd64 linux_386; do \
		GOOS="$${oa%_*}" GOARCH="$${oa#*_}" go build -o release/files/gitprompt cmd/gitprompt/gitprompt.go; \
		tar -C release/files -czf release/gitprompt_$${r}_$${oa}.tar.gz .; \
	done
	rm -rf release/files
	cd release && sha256sum gitprompt_*.tar.gz >checksums.txt
	cat release/checksums.txt


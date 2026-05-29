BIN := bin/aiguard

.PHONY: build test lint smoke clean

build:
	go build -o $(BIN) ./cmd/aiguard

test:
	go test ./...

lint:
	go vet ./...

smoke:
	./scripts/smoke.sh

clean:
	rm -rf $(BIN) .smoke
	go clean -testcache

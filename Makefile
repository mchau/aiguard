BIN := bin/aiguard

.PHONY: build test lint clean

build:
	go build -o $(BIN) ./cmd/aiguard

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f $(BIN)
	go clean -testcache

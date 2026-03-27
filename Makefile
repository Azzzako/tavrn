.PHONY: build run test clean

build:
	go build -o bin/tavrn ./cmd/tavrn
	go build -o bin/tavrn-admin ./cmd/tavrn-admin

run: build
	./bin/tavrn-admin

test:
	go test ./... -v

clean:
	rm -rf bin/ tavrn.db

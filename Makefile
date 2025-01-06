SOURCES := $(wildcard cmd/*.go cmd/*/*.go)

run: 
	go run ./cmd/main.go

got: $(SOURCES)
	$(eval GIT_TAG = $(shell git describe --tags))

	go build -ldflags="-X main.version=$(GIT_TAG)" -o build/got

test:
	go test ./...

testv:
	go test ./... -v

cover:
	go test ./... -coverprofile=cover.out
	go tool cover -html=cover.out
	rm cover.out
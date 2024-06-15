SOURCES := $(wildcard cmd/*.go cmd/*/*.go)

run: 
	go run ./cmd/main.go

got: $(SOURCES)
	go build -o build/got
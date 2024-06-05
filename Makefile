SOURCES := $(wildcard cmd/*.go cmd/*/*.go)

got: $(SOURCES)
	go build -o got cmd/main.go
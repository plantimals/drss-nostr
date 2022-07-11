all: build run

.PHONY: build

build:
	go build -o ./drsstools cmd/drsstools/drsstools.go

run:
	./drsstools

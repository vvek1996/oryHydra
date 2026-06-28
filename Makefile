.PHONY: install-deps code-build

install-deps:
	go mod download -C one
	go mod download -C two

code-build:
	go build -C one -o ../server .
	go build -C two -o ../test1 .

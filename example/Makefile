Bin="example"

.PHONY: proto run test build lint


default: build lint


build: proto
	go build -o $(Bin)


proto:
	@echo "Generating protobuf"
	@echo $(SHELL)
	@protoc --version
	@protoc proto/*.proto -I .  --proto_path=/usr/local/include/   --go_out=. --go_opt=paths=source_relative --myhttp_out=. --myhttp_opt=paths=source_relative --openapi_out=./docs/ --openapi_opt=naming=proto


run:build
	./$(Bin)

test:
	go test -race ./...

lint:
	golangci-lint run --fix
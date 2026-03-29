Bin="example"

.PHONY: proto run test build lint


default: build lint


build: proto
	go build -o $(Bin)


proto:
	@echo "Generating protobuf"
	@echo $(SHELL)
	@protoc --version
	@protoc proto/*.proto -I .  --proto_path=/usr/local/include/ --proto_path=$(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis/   --go_out=. --go_opt=paths=source_relative --myhttp_out=. --myhttp_opt=paths=source_relative --openapiv2_out=./docs/


run:build
	./$(Bin)

test:
	go test -race ./...

lint:
	golangci-lint run --fix
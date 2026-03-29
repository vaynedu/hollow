Bin="hollow"

.PHONY: build cli example run clean test

build:
	go build -o $(Bin) hollow.go

test:
	go test -v ./internal/...


fmt:
	go fmt ./...

lint:
	golangci-lint run --fix

proto:
	@echo "build proto"
	@cd ./example && pwd
	@protoc proto/*.proto --openapiv2_out=./docs/ --openapiv2_opt=naming=proto \
		--go_out=. --go_opt=paths=source_relative \
		--grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
		--validate_out=. --validate_opt=paths=source_relative

#$ protoc proto/*.proto -I .   --proto_path=/usr/local/include/   --go_out=. --go_opt=paths=source_relative      --myhttp_out=.
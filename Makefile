Bin="hollow"

build:
	go build -o $(Bin)

lint:
	golangci-lint run --fix
proto:
	@echo "build proto"
	@cd ./example && pwd
	@protoc proto/*.proto --openapi_out=./docs/ --openapi_opt=naming=proto \
    		--go_out=. --go_opt=paths=source_relative \
    		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
    		--http_out=. --http_opt=paths=source_relative \
    		--validate_out=. --validate_opt=paths=source_relative
	@protoc proto/*.proto --gotag_out=. --gotag_opt=paths=source_relative

#$ protoc proto/*.proto -I .   --proto_path=/usr/local/include/   --go_out=. --go_opt=paths=source_relative      --myhttp_out=.
Bin="example"

# 检测 hollow-cli 路径
HOLLOW_CLI := $(shell which hollow-cli 2>/dev/null || echo "../cmd/hollow_cli/hollow-cli")

.PHONY: proto run test build lint


default: build lint


build: proto
	go build -o $(Bin)


proto:
	@echo "🚀 Generating protobuf..."
	@protoc --version
	@protoc proto/*.proto -I .  --proto_path=/usr/local/include/ --proto_path=$(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis/   --go_out=. --go_opt=paths=source_relative --myhttp_out=. --myhttp_opt=paths=source_relative --openapiv2_out=./docs/
	@echo "🚀 Generating Hollow framework code..."
	@echo "   Using hollow-cli: $(HOLLOW_CLI)"
	@for proto_file in proto/*.proto; do \
		echo "   Processing: $$proto_file"; \
		$(HOLLOW_CLI) proto $$proto_file -f; \
	done
	@echo "✅ Code generation complete!"


run:build
	./$(Bin)

test:
	go test -race ./...

lint:
	golangci-lint run --fix

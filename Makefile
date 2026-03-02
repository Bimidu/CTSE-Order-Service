.PHONY: proto build run test lint docker-build docker-run

PROTO_DIR=proto
BINARY=order-service

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       $(PROTO_DIR)/auth/auth.proto \
	       $(PROTO_DIR)/product/product.proto \
	       $(PROTO_DIR)/order/order.proto

build:
	go build -o bin/$(BINARY) ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	go vet ./...

docker-build:
	docker build -t ctse-order-service .

docker-run:
	docker-compose up --build

tidy:
	go mod tidy

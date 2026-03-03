# ---- Proto generation stage ----
FROM golang:1.25-alpine AS proto-builder
RUN apk add --no-cache protobuf git
# Pin protoc plugins to versions compatible with Go 1.23
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY proto/ proto/
RUN protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           proto/auth/auth.proto proto/product/product.proto proto/order/order.proto

# ---- Build stage ----
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY --from=proto-builder /app/go.mod /app/go.sum ./
RUN go mod download
COPY --from=proto-builder /app/proto ./proto
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /order-service ./cmd/server

# ---- Runtime stage ----
FROM gcr.io/distroless/static-debian12
COPY --from=builder /order-service /order-service
EXPOSE 6969 50052
ENTRYPOINT ["/order-service"]

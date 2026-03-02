# CTSE Order Service

A production-ready Order microservice for the Rule-Based Mood-Driven E-Commerce Platform, built with **Go + Gin** and **PostgreSQL**.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                   API Gateway (Nginx)                │
└────────────┬───────────────────────┬────────────────┘
             │ REST /api/v1/*        │ REST /auth/*
             ▼                       ▼
   ┌──────────────────┐    ┌──────────────────┐
   │   Order Service  │    │   Auth Service   │
   │  :6969 (HTTP)    │◄───│  :4000 (HTTP)    │
   │  :50052 (gRPC)   │    │  :50051 (gRPC)   │
   └────────┬─────────┘    └──────────────────┘
            │ gRPC
            ▼
   ┌──────────────────┐
   │  Product Service │
   │  :50053 (gRPC)   │
   └──────────────────┘
```

## Technology Stack

| Component | Technology |
|-----------|-----------|
| Language  | Go 1.22   |
| Framework | Gin       |
| Database  | PostgreSQL (Neon — serverless) |
| ORM       | GORM      |
| Inter-service | gRPC (Protocol Buffers) |
| Auth      | JWT (local verify + Auth Service gRPC) |
| CI/CD     | GitHub Actions → GHCR → AWS ECS |

---

## Database Setup (Neon PostgreSQL — Free Tier)

1. Go to [neon.tech](https://neon.tech) and sign up with GitHub.
2. Click **New Project** → name it `ctse-order-service` → choose region `us-east-1` (closest to your AWS deployment).
3. Once created, find the **Connection string** on the dashboard. It looks like:
   ```
   postgresql://user:password@ep-xxx-yyy.us-east-1.aws.neon.tech/neondb?sslmode=require
   ```
4. Copy this into your `.env` as `DATABASE_URL`.

> GORM will auto-migrate all tables on first startup — no manual SQL needed.

---

## Local Development

### Prerequisites
- Go 1.22+
- `protoc` + Go plugins (run `make proto` to regenerate proto files)

### Setup

```bash
cp .env.example .env
# Edit .env — fill in DATABASE_URL, JWT_SECRET (must match Auth Service)
```

```bash
make run
```

The service starts on:
- **HTTP**: `http://localhost:6969`
- **gRPC**: `localhost:50052`

### Run tests

```bash
make test
```

---

## Docker

```bash
# Build and run with Docker Compose
make docker-run

# Build image only
make docker-build
```

---

## API Reference

All endpoints (except `/health`) require `Authorization: Bearer <token>`.

### Health
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/health` | None | Service health check |

### Cart
| Method | Path | Role | Description |
|--------|------|------|-------------|
| `POST` | `/api/v1/cart/items` | user | Add item to cart |
| `GET` | `/api/v1/cart` | user | Get current cart |
| `PUT` | `/api/v1/cart/items/:id` | user | Update item quantity |
| `DELETE` | `/api/v1/cart/items/:id` | user | Remove item |
| `DELETE` | `/api/v1/cart` | user | Clear cart |

### Orders
| Method | Path | Role | Description |
|--------|------|------|-------------|
| `POST` | `/api/v1/orders/checkout` | user | Checkout (creates order from cart) |
| `GET` | `/api/v1/orders` | user | Get my orders (paginated) |
| `GET` | `/api/v1/orders/:id` | user | Get a specific order |

### Admin
| Method | Path | Role | Description |
|--------|------|------|-------------|
| `GET` | `/api/v1/admin/orders` | admin | Get all orders (paginated, filterable) |
| `PUT` | `/api/v1/admin/orders/:id/status` | admin | Update order status |

---

## gRPC Server (port 50052)

Exposed for other services to consume:

```protobuf
service OrderService {
  rpc GetOrder(GetOrderRequest) returns (OrderResponse);
  rpc GetUserOrders(GetUserOrdersRequest) returns (GetUserOrdersResponse);
  rpc GetUserOrderCount(GetUserOrderCountRequest) returns (GetUserOrderCountResponse);
}
```

---

## Inter-Service Communication

| Direction | Protocol | Purpose |
|-----------|----------|---------|
| Order → Auth | gRPC (`:50051`) | Verify JWT, get user info, increment purchase count |
| Order → Product | gRPC (`:50053`) | Validate stock availability, reduce stock after checkout |
| Other services → Order | gRPC (`:50052`) | Get order details, order count |

---

## CI/CD Pipeline

```
push to main
    │
    ├── test (go test, go vet)
    ├── sonar (SonarCloud SAST)
    ├── build-and-push (Docker → ghcr.io)
    └── deploy (AWS ECS Fargate)
```

### Required GitHub Secrets

| Secret | Description |
|--------|-------------|
| `SONAR_TOKEN` | SonarCloud project token |
| `AWS_ACCESS_KEY_ID` | AWS IAM user key |
| `AWS_SECRET_ACCESS_KEY` | AWS IAM secret |
| `AWS_REGION` | e.g. `us-east-1` |

### Required GitHub Environment Variables (production environment)
Set in repository → Settings → Environments → production.

---

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `PORT` | No | HTTP port (default: `6969`) |
| `GRPC_PORT` | No | gRPC port (default: `50052`) |
| `DATABASE_URL` | **Yes** | Neon PostgreSQL connection string |
| `JWT_SECRET` | **Yes** | Must match Auth Service `JWT_SECRET` |
| `AUTH_SERVICE_ADDR` | No | Auth gRPC address (default: `auth-service:50051`) |
| `PRODUCT_SERVICE_ADDR` | No | Product gRPC address (default: `product-service:50053`) |
| `APP_ENV` | No | `development` or `production` |

---

## Regenerating Proto Files

```bash
make proto
```

This uses `protoc` with `protoc-gen-go` and `protoc-gen-go-grpc` plugins. Install with:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

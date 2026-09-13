<div align="center">

# Axis 🚀

**A Modern Personal Finance & Productivity Workspace**

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![Code Style](https://img.shields.io/badge/code%20style-standard-brightgreen.svg)](https://golang.org/doc/effective_go.html)

[English](README.md) · [简体中文](README_ZH.md)

</div>

---

## 📖 About

**Axis** is a comprehensive microservices-based personal workspace platform built with Go, designed to help you manage your financial life, productivity, and digital assets in one unified system.

### 🎯 Vision

Create a private, secure, and extensible platform that integrates financial management, productivity tools, and Web3 capabilities, empowering individuals to take full control of their digital financial life.

---

## ✨ Features

### 💰 Financial Management

- [ ] **Accounting & Bookkeeping**
  - [ ] Multi-currency support
  - [ ] Category management
  - [ ] Transaction tracking
  - [ ] Receipt upload & OCR
  
- [ ] **Investment Management**
  - [ ] Portfolio tracking
  - [ ] Stock market integration
  - [ ] Cryptocurrency tracking
  - [ ] Real-time price updates
  - [ ] P&L analysis
  
- [ ] **Wealth Analytics**
  - [ ] Net worth tracking
  - [ ] Asset allocation visualization
  - [ ] Performance reports
  - [ ] Tax optimization suggestions

### 📈 Market Data

- [ ] **Stock Market**
  - [ ] Real-time quotes
  - [ ] Historical data
  - [ ] Technical indicators
  - [ ] Watchlist management
  
- [ ] **Cryptocurrency**
  - [ ] Multi-exchange support
  - [ ] DeFi protocol integration
  - [ ] Price alerts
  - [ ] Market sentiment analysis

### 🗓️ Productivity

- [ ] **Calendar & Scheduling**
  - [ ] Event management
  - [ ] Reminders & notifications
  - [ ] Google Calendar sync
  - [ ] Time zone support
  
- [ ] **Task Management**
  - [ ] Todo lists
  - [ ] Project tracking
  - [ ] Priority management
  - [ ] Deadline alerts

### 🕷️ Automation

- [ ] **Web Scraping Workflows**
  - [ ] Visual workflow designer
  - [ ] Scheduled tasks
  - [ ] Data extraction rules
  - [ ] Output formatting
  
- [ ] **Custom Integrations**
  - [ ] Webhook support
  - [ ] API connectors
  - [ ] Data pipelines
  - [ ] Event-driven automation

### 🌐 Web3 Integration

- [ ] **Wallet Management**
  - [ ] HD wallet creation
  - [ ] Multi-chain support (ETH, BTC, Solana, etc.)
  - [ ] Private key encryption
  - [ ] Hardware wallet integration
  
- [ ] **Blockchain Interactions**
  - [ ] Transaction signing & broadcasting
  - [ ] Smart contract interaction
  - [ ] Token transfers
  - [ ] Gas optimization
  
- [ ] **On-Chain Monitoring**
  - [ ] Address tracking
  - [ ] Transaction alerts
  - [ ] Balance monitoring
  - [ ] DeFi position tracking

---

## 🏗️ Architecture

Axis follows a **microservices architecture** with **Domain-Driven Design (DDD)** and **CQRS** patterns.

### Core Services

```
┌─────────────────────────────────────────────────────────────┐
│                        API Gateway                           │
│                   (Routing, Auth, Rate Limiting)             │
└────────────┬────────────────────────────────┬───────────────┘
             │                                │
    ┌────────▼────────┐              ┌───────▼────────┐
    │  Auth Service   │              │  User Service  │
    │  (JWT, RBAC)    │              │  (Profile)     │
    └────────┬────────┘              └───────┬────────┘
             │                                │
    ┌────────▼────────────────────────────────▼────────┐
    │              Event Bus (Kafka)                    │
    └────────┬────────────────────────────┬─────────────┘
             │                            │
    ┌────────▼────────┐          ┌───────▼──────────┐
    │ Wallet Service  │          │ Ledger Service   │
    │ (Crypto/Stocks) │          │ (Bookkeeping)    │
    └─────────────────┘          └──────────────────┘
             │                            │
    ┌────────▼────────┐          ┌───────▼──────────┐
    │ Notification    │          │ Audit Service    │
    │ Service         │          │ (Logging)        │
    └─────────────────┘          └──────────────────┘
```

### Technology Stack

| Category | Technology |
|----------|-----------|
| **Language** | Go 1.26+ |
| **Framework** | Gin, gRPC, Protocol Buffers |
| **Database** | MySQL, Redis, MongoDB |
| **Message Queue** | Kafka |
| **Service Discovery** | Consul |
| **Monitoring** | Prometheus, Grafana, Jaeger |
| **Container** | Docker, Kubernetes |
| **CI/CD** | GitHub Actions |

### Design Patterns

- ✅ **Domain-Driven Design (DDD)**
- ✅ **CQRS (Command Query Responsibility Segregation)**
- ✅ **Event Sourcing** (for audit trail)
- ✅ **Saga Pattern** (for distributed transactions)
- ✅ **Outbox Pattern** (for event reliability)
- ✅ **Repository Pattern**
- ✅ **Decorator Pattern** (for cross-cutting concerns)
- ✅ **Dependency Injection** (using Wire)

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.26+**
- **Docker & Docker Compose**
- **MySQL 8.0+**
- **Redis 7.0+**
- **Kafka 3.0+** (optional, for async messaging)
- **Consul** (optional, for service discovery)

### Installation

1. **Clone the repository**

```bash
git clone https://github.com/Crows-Storm/Axis.git
cd Axis
```

2. **Start infrastructure services**

```bash
docker-compose up -d mysql redis consul
```

3. **Run database migrations**

```bash
make migrate-up
```

4. **Start services**

```bash
# Start all services
make run-all

# Or start individual services
make run-auth
make run-user
make run-wallet
```

5. **Access the API**

```bash
# Health check
curl http://localhost:8000/health

# API documentation
open http://localhost:8000/swagger
```

### Development Setup

```bash
# Install dependencies
make deps

# Generate protobuf files
make proto

# Generate OpenAPI clients
make openapi

# Run tests
make test

# Run linter
make lint

# Build all services
make build
```

---

## 📁 Project Structure

```
Axis/
├── api/                        # API definitions
│   ├── proto/                  # gRPC proto files
│   └── openapi/                # OpenAPI/Swagger specs
│
├── services/                   # Microservices
│   ├── auth/                   # Authentication & Authorization
│   ├── user/                   # User management
│   ├── wallet/                 # Wallet & portfolio
│   ├── ledger/                 # Bookkeeping & accounting
│   ├── audit/                  # Audit logging
│   ├── notification/           # Notifications
│   └── gateway/                # API gateway
│
├── common/                     # Shared libraries
│   ├── domain/                 # Domain primitives
│   ├── decorator/              # CQRS decorators
│   ├── server/                 # Server utilities
│   └── client/                 # Service clients
│
├── pkg/                        # Public libraries
│   ├── logger/                 # Logging
│   ├── metrics/                # Monitoring
│   ├── discovery/              # Service discovery
│   └── security/               # Security utilities
│
├── deployments/                # Deployment configs
│   ├── docker/                 # Dockerfiles
│   └── kubernetes/             # K8s manifests
│
├── scripts/                    # Build & deployment scripts
├── docs/                       # Documentation
├── test/                       # Integration & E2E tests
│
├── docker-compose.yml          # Local development
├── Makefile                    # Build automation
├── go.work                     # Go workspace
└── README.md                   # This file
```

### Service Structure (DDD Layered Architecture)

```
service/
├── cmd/
│   └── api/                    # Application entry point
│       ├── main.go
│       └── wire.go             # Dependency injection
│
├── internal/                   # Private code
│   ├── domain/                 # Domain layer (business logic)
│   │   ├── user/
│   │   │   ├── user.go         # Aggregate root
│   │   │   ├── repository.go   # Repository interface
│   │   │   └── events.go       # Domain events
│   │
│   ├── application/            # Application layer (use cases)
│   │   ├── command/            # Commands (write)
│   │   ├── query/              # Queries (read)
│   │   └── service/            # Application services
│   │
│   ├── infrastructure/         # Infrastructure layer
│   │   ├── persistence/        # Database implementations
│   │   ├── messaging/          # Message queue
│   │   └── external/           # External service clients
│   │
│   └── ports/                  # Ports layer (interfaces)
│       ├── http/               # HTTP handlers
│       ├── grpc/               # gRPC services
│       └── events/             # Event consumers
│
├── configs/                    # Configuration files
├── go.mod
└── Makefile
```

---

## 🧪 Testing

```bash
# Run unit tests
make test

# Run integration tests
make test-integration

# Run E2E tests
make test-e2e

# Generate coverage report
make coverage

# Run specific service tests
cd services/auth && go test ./...
```

---

## 📚 Documentation

- [Architecture Overview](docs/architecture-review.md)
- [Domain Event Design](docs/domain-event-design.md)
- [API Documentation](docs/api/)
- [Deployment Guide](docs/deployment/)
- [Development Guide](docs/development/)
- [Contributing Guidelines](CONTRIBUTING.md)

---

## 🛣️ Roadmap

### Phase 1: Foundation (Q1 2026) 🏗️

- [ ] Core microservices setup
- [ ] Authentication & authorization
- [ ] User management
- [ ] Basic API gateway
- [ ] Database schema design
- [ ] Docker containerization

### Phase 2: Financial Core (Q2 2026) 💰

- [ ] Ledger service (accounting)
- [ ] Transaction management
- [ ] Multi-currency support
- [ ] Basic reporting
- [ ] Category management
- [ ] Budget tracking

### Phase 3: Investment Tracking (Q3 2026) 📈

- [ ] Wallet service
- [ ] Stock market integration
- [ ] Cryptocurrency tracking
- [ ] Portfolio analytics
- [ ] Real-time price updates
- [ ] P&L calculations

### Phase 4: Automation & Web3 (Q4 2026) 🤖

- [ ] Web scraping workflows
- [ ] Scheduled tasks
- [ ] Web3 wallet creation
- [ ] Blockchain integrations
- [ ] DeFi tracking
- [ ] Smart contract interactions

### Phase 5: Advanced Features (2027+) 🚀

- [ ] Mobile app (React Native)
- [ ] Advanced analytics & ML
- [ ] Social features
- [ ] Plugin system
- [ ] White-label solution
- [ ] Enterprise features

---

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Standards

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Write unit tests for new features
- Update documentation
- Use conventional commits
- Run linters before submitting

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- [golang-standards/project-layout](https://github.com/golang-standards/project-layout) for project structure inspiration
- [Domain-Driven Design](https://www.domainlanguage.com/ddd/) by Eric Evans
- [Implementing Domain-Driven Design](https://vaughnvernon.com/) by Vaughn Vernon
- [Microservices Patterns](https://microservices.io/patterns/) by Chris Richardson

---

## 📞 Contact & Support

- **Issues**: [GitHub Issues](https://github.com/Crows-Storm/Axis/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Crows-Storm/Axis/discussions)
- **Email**: crow@axis-platform.com

---

## ⭐ Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Crows-Storm/Axis&type=Date)](https://star-history.com/#Crows-Storm/Axis&Date)

---

<div align="center">

**Made with ❤️ by [Crow](https://github.com/Crows-Storm)**

If you find this project helpful, please consider giving it a ⭐!

</div>

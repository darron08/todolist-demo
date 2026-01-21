# Todo List Demo

A high-performance, scalable todo list microservice built with Go and Clean Architecture principles.

## Features

- User authentication with JWT
- Todo CRUD operations
- Real-time collaboration
- Multi-environment support
- Containerized deployment

## Tech Stack

- **Language**: Go 1.24.0
- **Web Framework**: Gin
- **Database**: MySQL 8.0
- **Cache**: Redis
- **ORM**: GORM v2
- **Configuration**: Viper
- **Containerization**: Docker & Kubernetes

## Project Structure

```
.
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── domain/            # Domain layer (entities, repositories)
│   ├── usecase/           # Use case layer (business logic)
│   ├── infrastructure/    # Infrastructure layer (database, external services)
│   └── interfaces/        # Interface layer (HTTP handlers, websockets)
├── pkg/                   # Public library code
├── configs/               # Configuration files
├── migrations/            # Database migrations
├── deployments/           # Kubernetes configurations
└── scripts/               # Build and deployment scripts
```

## Getting Started

### Prerequisites

- Go 1.24.0+
- Docker & Docker Compose
- MySQL 8.0+
- Redis (optional)

### Installation

1. Clone the repository
2. Copy and configure environment files:
   ```bash
   cp configs/config.dev.yaml configs/config.local.yaml
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Run database migrations:
   ```bash
   make migrate-up
   ```
5. Start the application:
   ```bash
   make run
   ```

### Development

```bash
# Build the application
make build

# Run tests
make test

# Run with hot reload
make dev

# Run linter
make lint
```

## API Documentation

API documentation is available at `/swagger/index.html` when running the application.

## Contributing

Please follow the contribution guidelines in CONTRIBUTING.md.

## License

This project is licensed under the MIT License.
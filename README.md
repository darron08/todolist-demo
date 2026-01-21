# Todo List Demo

A high-performance, scalable todo list microservice built with Go and Clean Architecture principles.

## Features

- User authentication with JWT
- Todo CRUD operations
- Real-time collaboration
- Containerized deployment

## Tech Stack

- **Language**: Go 1.24.0
- **Web Framework**: Gin
- **Database**: MySQL 8.0
- **Cache**: Redis
- **ORM**: GORM v2
- **Configuration**: Viper
- **Containerization**: Docker & Docker Compose

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
└── tests/                 # Integration and unit tests
```

## Getting Started

### Quick Start (Docker Compose - 推荐)

Easiest way to start the application for demo or development:

1. Clone the repository:
   ```bash
   git clone https://github.com/darron08/todolist-demo.git
   cd todolist-demo
   ```

2. Start all services (API + MySQL + Redis):
   ```bash
   docker-compose up -d
   ```

3. Verify startup:
   ```bash
   docker-compose ps
   docker-compose logs -f api
   ```

4. Access the application:
   - API: http://localhost:8080
   - Swagger Documentation: http://localhost:8080/swagger/index.html
   - Health Check: http://localhost:8080/health

5. Default admin credentials:
   - Username: `admin`
   - Password: `admin`

6. Stop services:
   ```bash
   docker-compose down
   ```

### Local Development

#### Prerequisites

- Go 1.24.0+
- MySQL 8.0+ (running on localhost:3306)
- Redis (running on localhost:6379)

#### Setup Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/darron08/todolist-demo.git
   cd todolist-demo
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Start MySQL & Redis (choose one):

   **Option A: Use Docker Compose for dependencies only**
   ```bash
   docker-compose up -d mysql redis
   ```

   **Option B: Use local MySQL/Redis**
   ```bash
   # Ensure MySQL is running on localhost:3306
   # Ensure Redis is running on localhost:6379
   ```

4. Run database migrations:
   ```bash
   make migrate-up
   ```

5. Start the application:
   ```bash
   make run
   ```

6. Verify startup:
   ```bash
   curl http://localhost:8080/health
   ```

The API will be available at http://localhost:8080

### Development Commands

```bash
# Build the application
make build

# Run the application
make run

# Run tests
make test

# Run tests with coverage
make test-cover

# Run linter
make lint

# Format code
make fmt

# Security scan
make security

# Clean build artifacts
make clean

# Download dependencies
make deps

# Database migrations
make migrate-up      # Run migrations
make migrate-down    # Rollback migrations
make migrate-create NAME=add_new_table  # Create new migration
```

## Usage Examples

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

### Create Todo

```bash
curl -X POST http://localhost:8080/api/v1/todos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-access-token>" \
  -d '{
    "title": "Buy groceries",
    "description": "Milk, eggs, bread",
    "due_date": "2026-01-25T10:00:00Z",
    "priority": "high"
  }'
```

### Get All Todos

```bash
curl -X GET http://localhost:8080/api/v1/todos \
  -H "Authorization: Bearer <your-access-token>"
```

### Update Todo

```bash
curl -X PUT http://localhost:8080/api/v1/todos/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-access-token>" \
  -d '{
    "title": "Buy groceries (updated)",
    "completed": true
  }'
```

### Delete Todo

```bash
curl -X DELETE http://localhost:8080/api/v1/todos/1 \
  -H "Authorization: Bearer <your-access-token>"
```

## API Documentation

Interactive API documentation is available at `/swagger/index.html` when running the application:
- Swagger UI: http://localhost:8080/swagger/index.html

Full API specifications include authentication, todo management, and admin endpoints.

## Contributing

Please follow the contribution guidelines in CONTRIBUTING.md.

## License

This project is licensed under the MIT License.
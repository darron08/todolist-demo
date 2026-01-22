# Todo List Demo

A high-performance, scalable todo list microservice built with Go and Clean Architecture principles.

## Features

- User authentication with JWT
- Todo CRUD operations
- Real-time collaboration
- Containerized deployment (Docker & Docker Compose)

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

Interactive API documentation is available via Swagger UI when running the application:
- Swagger UI: http://localhost:8080/swagger/index.html
- Swagger JSON: http://localhost:8080/swagger/doc.json
- Swagger YAML: http://localhost:8080/swagger/doc.yaml

### Using Swagger UI

1. **Access the documentation**
   - Start the application: `docker-compose up -d` or `make run`
   - Open your browser and navigate to: http://localhost:8080/swagger/index.html

2. **Explore API endpoints**
   - The API is organized into tags: Authentication, Users, Todos, Admin, and Health
   - Click on any tag to expand and view its endpoints

3. **Test API endpoints**
   - Click on an endpoint to view its details (parameters, responses, etc.)
   - Click the **Try it out** button
   - Fill in the required parameters
   - Click **Execute** to send a real request
   - View the response, status code, and request duration

4. **Authentication**
   - For endpoints requiring authentication, you need to provide a Bearer token
   - First, use the `/api/v1/auth/login` endpoint to get your access token
   - Copy the `access_token` from the response
   - Click the **Authorize** button in Swagger UI (top right)
   - Enter your token with the `Bearer ` prefix: `Bearer <your-access-token>`
   - Click **Authorize** and close the dialog
   - Now you can call authenticated endpoints

### API Endpoint Overview

#### Authentication (Public)
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - User logout

#### Users (Requires Authentication)
- `GET /api/v1/users/profile` - Get current user profile

#### Todos (Requires Authentication)
- `POST /api/v1/todos` - Create a new todo
- `GET /api/v1/todos` - List todos (with pagination and filters)
- `GET /api/v1/todos/:id` - Get a specific todo
- `PUT /api/v1/todos/:id` - Update a todo
- `DELETE /api/v1/todos/:id` - Delete a todo
- `PATCH /api/v1/todos/:id/status` - Update todo status

#### Admin (Requires Admin Role)
- `POST /api/v1/admin/users` - Create a user
- `GET /api/v1/admin/users` - List all users
- `GET /api/v1/admin/users/:id` - Get a specific user
- `DELETE /api/v1/admin/users/:id` - Delete a user
- `GET /api/v1/admin/todos` - List all todos
- `DELETE /api/v1/admin/todos/:id` - Delete any todo

#### Health Checks
- `GET /health` - Health check
- `GET /ready` - Readiness check

### Regenerating Swagger Documentation

If you modify any API endpoints or handlers, regenerate the Swagger documentation:

```bash
make swagger
# or
swag init -g cmd/api/main.go -o docs
```

The documentation will be regenerated in the `docs/` directory.

## Contributing

Please follow the contribution guidelines in CONTRIBUTING.md.

## License

This project is licensed under the MIT License.
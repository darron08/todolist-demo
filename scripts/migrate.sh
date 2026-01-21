#!/bin/bash

# Database migration script for Todo List API

set -e

# Default values
MIGRATION_PATH=${MIGRATION_PATH:-"migrations"}
DATABASE_URL=${DATABASE_URL:-"mysql://root:password@tcp(localhost:3306)/todolist_demo"}
MIGRATE_TOOL=${MIGRATE_TOOL:-"migrate"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if migrate tool is installed
check_migrate_tool() {
    if ! command -v $MIGRATE_TOOL &> /dev/null; then
        print_error "migrate tool is not installed"
        print_status "Installing migrate tool..."
        go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
        
        if ! command -v $MIGRATE_TOOL &> /dev/null; then
            print_error "Failed to install migrate tool"
            exit 1
        fi
        print_status "migrate tool installed successfully"
    fi
}

# Function to validate migration path
validate_migration_path() {
    if [ ! -d "$MIGRATION_PATH" ]; then
        print_error "Migration path '$MIGRATION_PATH' does not exist"
        exit 1
    fi
    
    if [ ! "$(ls -A $MIGRATION_PATH)" ]; then
        print_warning "Migration path '$MIGRATION_PATH' is empty"
    fi
}

# Function to wait for database
wait_for_database() {
    print_status "Waiting for database connection..."
    
    # Extract host and port from DATABASE_URL
    HOST=$(echo $DATABASE_URL | sed -n 's/.*@\([^:]*\):.*/\1/p')
    PORT=$(echo $DATABASE_URL | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')
    
    if [ -z "$HOST" ]; then
        HOST="localhost"
    fi
    
    if [ -z "$PORT" ]; then
        PORT="3306"
    fi
    
    # Wait for database to be ready
    timeout 60 bash -c "until nc -z $HOST $PORT; do sleep 1; done" 2>/dev/null || \
    timeout 60 bash -c "until telnet $HOST $PORT </dev/null 2>/dev/null; do sleep 1; done" || \
    {
        print_warning "Could not verify database connection, proceeding anyway..."
    }
    
    print_status "Database connection assumed ready"
}

# Function to run migrations up
migrate_up() {
    print_status "Running database migrations up..."
    $MIGRATE_TOOL -path "$MIGRATION_PATH" -database "$DATABASE_URL" up
    print_status "Migrations completed successfully"
}

# Function to run migrations down
migrate_down() {
    if [ -z "$STEPS" ]; then
        print_error "Please specify STEPS environment variable for migration down"
        print_status "Usage: STEPS=1 ./scripts/migrate.sh down"
        exit 1
    fi
    
    print_status "Running database migrations down ($STEPS steps)..."
    $MIGRATE_TOOL -path "$MIGRATION_PATH" -database "$DATABASE_URL" down $STEPS
    print_status "Migration rollback completed successfully"
}

# Function to get migration version
migrate_version() {
    print_status "Getting current migration version..."
    $MIGRATE_TOOL -path "$MIGRATION_PATH" -database "$DATABASE_URL" version
}

# Function to force migration version
migrate_force() {
    if [ -z "$VERSION" ]; then
        print_error "Please specify VERSION environment variable for force"
        print_status "Usage: VERSION=1 ./scripts/migrate.sh force"
        exit 1
    fi
    
    print_status "Force setting migration version to $VERSION..."
    $MIGRATE_TOOL -path "$MIGRATION_PATH" -database "$DATABASE_URL" force $VERSION
    print_status "Migration version forced to $VERSION"
}

# Function to create new migration
create_migration() {
    if [ -z "$NAME" ]; then
        print_error "Please specify NAME environment variable for migration creation"
        print_status "Usage: NAME=create_user_table ./scripts/migrate.sh create"
        exit 1
    fi
    
    print_status "Creating new migration: $NAME"
    $MIGRATE_TOOL create -ext sql -dir "$MIGRATION_PATH" -seq "$NAME"
    print_status "Migration files created successfully"
}

# Main script logic
case "${1:-up}" in
    "up")
        check_migrate_tool
        validate_migration_path
        wait_for_database
        migrate_up
        ;;
    "down")
        check_migrate_tool
        validate_migration_path
        wait_for_database
        migrate_down
        ;;
    "version")
        check_migrate_tool
        validate_migration_path
        migrate_version
        ;;
    "force")
        check_migrate_tool
        validate_migration_path
        migrate_force
        ;;
    "create")
        check_migrate_tool
        create_migration
        ;;
    "help"|"-h"|"--help")
        echo "Database Migration Script"
        echo ""
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  up       Run all pending migrations (default)"
        echo "  down     Rollback migrations (requires STEPS env var)"
        echo "  version  Get current migration version"
        echo "  force    Force set migration version (requires VERSION env var)"
        echo "  create   Create new migration files (requires NAME env var)"
        echo "  help     Show this help message"
        echo ""
        echo "Environment Variables:"
        echo "  MIGRATION_PATH  Path to migration files (default: migrations)"
        echo "  DATABASE_URL    Database connection URL"
        echo "  STEPS           Number of migrations to rollback (for down command)"
        echo "  VERSION         Migration version to force set (for force command)"
        echo "  NAME            Migration name (for create command)"
        ;;
    *)
        print_error "Unknown command: $1"
        print_status "Use '$0 help' for usage information"
        exit 1
        ;;
esac
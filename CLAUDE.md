# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Building

```bash
make build           # Build the yamlforge binary
go build -o yamlforge cmd/yamlforge/main.go  # Alternative direct build
```

### Testing

```bash
make test            # Run all tests
go test ./...        # Alternative direct test
go test -v ./internal/parser  # Run specific package tests
```

### Running Examples

```bash
make run-blog        # Run blog example on port 8080
make run-tasks       # Run tasks example on port 8080
./yamlforge serve examples/blog.yaml -port 3000  # Custom port
```

### Validation

```bash
make validate        # Validate all example YAML files
./yamlforge validate examples/blog.yaml  # Validate specific config
```

### Development

```bash
make dev             # Build and run blog example
make clean           # Clean build artifacts
make deps            # Download dependencies
```

## Architecture

### Application Structure

YamlForge is a single-binary web application generator that creates complete web apps from YAML configurations. It follows a modular architecture with clear separation of concerns:

1. **Entry Point** (`cmd/yamlforge/main.go`): CLI interface handling commands (serve, validate) and flags
2. **Configuration Parser** (`internal/parser/`): Parses and validates YAML configurations into strongly-typed Go structs
3. **Database Layer** (`internal/database/`): Factory pattern for database backends (SQLite, PostgreSQL, MySQL)
4. **API Layer** (`internal/api/`): RESTful handlers with automatic CRUD operations for each model
5. **Server** (`internal/server/`): HTTP server setup with routing, CORS, and middleware
6. **UI** (`internal/ui/`): Embedded HTML templates and static assets for the web interface
7. **Validation** (`internal/validation/`): Field-level validation logic based on YAML rules

### Key Design Patterns

- **Factory Pattern**: Database creation abstracts different DB types through a common interface
- **Template-Based UI**: HTML templates are embedded in the binary for zero-dependency deployment
- **Dynamic Schema Generation**: Models defined in YAML automatically generate database schemas and API endpoints
- **Single Binary Distribution**: All assets and templates are embedded, requiring no external files

### Request Flow

1. YAML config defines models with fields, validations, and UI preferences
2. Server initializes database schema based on models
3. API handlers are dynamically created for each model
4. UI templates render forms and lists based on model definitions
5. All CRUD operations go through validation layer before database operations

### Model Configuration

Models support various field types (text, number, boolean, datetime, enum, relation) with validations (required, unique, min/max, pattern). Each model can have custom UI configurations for list/form views and permission settings.


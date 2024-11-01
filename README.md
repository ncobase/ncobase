# Ncobase

## Start

```shell
# Install dependencies
go mod tidy

# Generate necessary files
make generate

# Run the application
make run
```

## Technologies

[Golang](https://go.dev), [PostgreSQL](https://www.postgresql.org) / [MySQL](https://www.mysql.com), [Gin](https://github.com/gin-gonic/gin), [ent.](https://entgo.io), [Swagger 2.0](https://github.com/swaggo/gin-swagger)

## Project structure

```plaintext.
├── bin                  # Compiled executable files
│   └── plugins          # Plugin binaries
├── cmd                  # Command-line applications
│   ├── cli              # CLI tools and utilities
│   └── ncobase          # Main program entry
├── core                 # Core system components
│   ├── access           # Access control and permissions
│   ├── auth             # Authentication and authorization
│   ├── space            # Workspace management
│   ├── system           # System-wide functionalities
│   ├── tenant           # Multi-tenancy support
│   └── user             # User management
├── docs                 # Documentation
│   └── swagger          # API documentation
├── domain               # Business domain logic
│   ├── content          # Content management
│   ├── realtime         # Real-time functionality
│   └── resource         # Resource management
├── front                # Frontend codebase
│   ├── apps             # Frontend applications
│   ├── infra            # Frontend infrastructure
│   ├── packages         # Shared frontend packages
│   └── plugins          # Frontend plugins
├── infra                # Infrastructure components
│   ├── config           # Configuration files
│   ├── database         # Database configurations
│   └── systemd          # Systemd service files
├── logs                 # Log files
├── pkg                  # Shared packages and utilities
│   ├── biz              # Business logic utilities
│   ├── cache            # Caching functionality
│   ├── config           # Configuration utilities
│   ├── consts           # Constants
│   ├── cookie           # Cookie management
│   ├── crypto           # Cryptography utilities
│   ├── data             # Data handling
│   ├── ecode            # Error codes
│   ├── elastic          # Elasticsearch integration
│   ├── email            # Email functionality
│   ├── entgo            # Ent ORM utilities
│   ├── feature          # Feature management
│   ├── helper           # Helper functions
│   ├── jwt              # JWT authentication
│   ├── log              # Logging utilities
│   ├── meili            # Meilisearch integration
│   ├── nanoid           # Unique ID generation
│   ├── oauth            # OAuth implementation
│   ├── observes         # Observability tools
│   ├── paging           # Pagination utilities
│   ├── proxy            # Proxy functionality
│   ├── resp             # Response handling
│   ├── router           # Routing utilities
│   ├── slug             # URL slug generation
│   ├── storage          # Storage management
│   ├── time             # Time utilities
│   ├── types            # Type definitions
│   ├── util             # General utilities
│   ├── uuid             # UUID generation
│   └── validator        # Data validation
├── plugin               # Plugin system
│   ├── counter          # Counter plugin
│   └── sample           # Sample plugin template
└── proxy                # API proxy functionality
```

## Documentation

For full documentation, including API references and deployment guides, visit [https://docs.nocobase.com](https://docs.nocobase.com).

## Maintainers

[@Shen](https://github.com/haiyon)

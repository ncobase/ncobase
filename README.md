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
├── bin                     # Compiled executable files
│   └── plugins             # Plugin binaries
├── cmd
│   └── ncobase             # Main program entry
├── docs                    # Documentation
├── feature                 # Core features and modules
│   ├── access              # Access control and permissions
│   ├── auth                # Authentication and authorization
│   ├── content             # Content management
│   ├── group               # User group management
│   ├── socket              # Data relationship management
│   ├── resource            # Resource management
│   ├── system              # System-wide functionalities
│   ├── tenant              # Multi-tenancy support
│   └── user                # User management
├── helper                  # Helper utilities and functions
├── infra                   # Infrastructure configurations
│   ├── config              # Configuration files
│   └── systemd             # Systemd service files
├── logs                    # Log files
├── pkg                     # Public packages
└── proxy                   # API proxy functionality
```

## Documentation

For full documentation, including API references and deployment guides, visit https://docs.nocobase.com.

## Maintainers

[@Shen](https://github.com/haiyon)

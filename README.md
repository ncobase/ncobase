# Ncobase

## Quick Start

```shell
# Setup
go mod tidy
go work sync
make install           # Install required tools (swag, etc.)

# Development
make generate         # Generate code and swagger docs
make swagger          # Generate swagger documentation
make run              # Run the application locally

# Build
make build            # Build for current platform
make build-multi      # Build for multiple platforms (linux/darwin)
make build-plugin     # Build plugin for current platform
make build-plugins    # Build plugins for all platforms
make build-business   # Build business extensions
make build-all        # Build application and all extensions

# Utils
make clean            # Clean build artifacts
make version          # Show version information
make help             # Show make commands help
```

## Technologies

[Golang](https://go.dev), [PostgreSQL](https://www.postgresql.org) / [MySQL](https://www.mysql.com), [Gin](https://github.com/gin-gonic/gin), [ent.](https://entgo.io), [Swagger 2.0](https://github.com/swaggo/gin-swagger)

## Documentation

- [Overview](docs/Overview.md)
- [Extension Design Development Guide](docs/Extension_Design_Development_Guide.md)
- [Business Domain Reference](docs/Business_Domain_Reference.md)

For full documentation, including API references and deployment guides,
visit [https://docs.nocobase.com](https://docs.nocobase.com).

## Maintainers

[@Shen](https://github.com/haiyon)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

# Ncobase

Business backend application for the Ncobase platform.

For the current cross-project plan and rules, see the root
[`PROJECT_PLAN.md`](../PROJECT_PLAN.md).

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

- [Overview](docs/OVERVIEW.md)
- [Domain Reference](docs/DOMAIN_REFERENCE.md)
- [API Contract Baseline](docs/API_CONTRACT.md)
- [Permission Rules](docs/PERMISSIONS.md)
- [Feature Interactions](docs/FEATURE_INTERACTIONS.md)
- [State Machines](docs/STATE_MACHINES.md)
- [Extension Guide](docs/EXTENSION_GUIDE.md)
- [Migration Notes](docs/MIGRATION_NOTES.md)
- [Swagger JSON](docs/swagger/swagger.json)
- [Archived Backend Documents](docs/archive/README.md)

For full documentation, including API references and deployment guides,
visit [https://docs.ncobase.com](https://docs.ncobase.com).

## Maintainers

[@Shen](https://github.com/haiyon)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

# Ncobase backend

## Start

```shell
# go mod
go mod tidy

# generate
make generate

# run
make run
```

## Technologies

[Golang](https://go.dev), [PostgreSQL](https://www.postgresql.org) / [MySQL](https://www.mysql.com), [Gin](https://github.com/gin-gonic/gin), [ent.](https://entgo.io), [GraphQL](https://graphql.org), [Swagger 2.0](https://github.com/swaggo/gin-swagger)

## Project structure

```plaintext
├── bin                     # Compiled executable files
│   └── plugins             # Plugin binaries
├── cmd
│   ├── bootstrap           # Bootstrap commands
│   └── ncobase             # Main program entry
├── core                    # Core application logic
│   ├── data                # Data handling
│   │   ├── ent             # ent ORM related
│   │   ├── graphql         # GraphQL schemas and generated code
│   │   ├── repository      # Repositories for data access
│   │   ├── schema          # Database schemas
│   │   └── structs         # Data structures
│   ├── graphql             # GraphQL resolvers and types
│   │   ├── generated       # Auto-generated GraphQL code
│   │   ├── resolvers       # GraphQL resolvers
│   │   └── types           # GraphQL type definitions
│   ├── handler             # Request handlers
│   └── service             # Business logic services
├── docs                    # Documentation
├── helper                  # Helper utilities and functions
├── infra                   # Infrastructure configurations
│   ├── config              # Configuration files
│   └── systemd             # Systemd service files
├── logs                    # Log files
├── middleware              # Middleware
├── pkg                     # Public packages
└── plugin                  # Plugin management
    ├── asset               # Asset plugin
    ├── content             # Content plugin
    └── user                # User plugin
```

## Documentation

For full documentation, visit [https://domain.com](https://domain.com).

## Maintainers

[@Shen](https://github.com/haiyon)

## License

[MIT](LICENSE)

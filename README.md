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
├── cmd
│   └── ncobase             # Main program entry
├── docs                    # Documentation
├── infra                   # Infrastructure configurations
│   ├── docker              # Docker configuration files
│   ├── kubernetes          # Kubernetes configuration files
│   └── systemd             # Systemd service files
├── internal                # Internal application logic
│   ├── app                 # Application layer containing business logic
│   │   ├── auth            # Authentication and authorization logic
│   │   ├── user            # User-related business logic
│   │   ├── file            # File management business logic
│   │   └── ...             # Other business modules
│   ├── config              # Configuration files
│   ├── data                # Data handling
│   │   ├── ent             # ent ORM related
│   │   ├── repository      # Repositories for data access
│   │   ├── schema          # Database schemas
│   │   └── structs         # Data structures
│   ├── graphql             # GraphQL resolvers and types
│   │   ├── generated       # Auto-generated GraphQL code
│   │   ├── resolvers       # GraphQL resolvers
│   │   └── types           # GraphQL type definitions
│   ├── handler             # Request handlers
│   ├── helper              # Helper utilities and functions
│   ├── middleware          # Middleware
│   ├── plugin              # Plugin management
│   │   ├── example         # Example plugin
│   │   │   ├── main.go     # Plugin main entry
│   │   │   ├── README.md   # Plugin documentation
│   │   │   └── plugin.json # Plugin configuration file
│   │   └── ...             # Other plugins
│   └── server              # Server-related code
│       └── middleware      # Middleware
│   └── service             # Business logic
└── logs                    # Log files
└── pkg                     # Public packages
    ├── cache               # Cache management
    ├── consts              # Constants
    ├── cookie              # Cookie handling
    ├── crypto              # Encryption utilities
    ├── ecode               # Error codes
    ├── elastic             # Elasticsearch support
    ├── email               # Email related logic
    ├── jwt                 # JWT handling
    ├── log                 # Logging
    ├── meili               # Meilisearch support
    ├── nanoid              # NanoID generation
    ├── oauth               # OAuth related logic
    ├── resp                # Response handling
    ├── slug                # Slug generation
    ├── storage             # Storage management
    ├── time                # Time utilities
    ├── types               # Type definitions
    ├── util                # Utility functions
    ├── uuid                # UUID generation
    └── validator           # Validators
└── scripts                 # Operational and management scripts
```

## Documentation

For full documentation, visit [https://domain.com](https://domain.com).

## Maintainers

[@Shen](https://github.com/haiyon)

## License

[MIT](LICENSE)

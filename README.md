# stocms backend

## Start

```shell
# go mod
go mod download

# generate
make generate

# run
make run
```

## Technologies

[Golang](https://go.dev), [PostgreSQL](https://www.postgresql.org) / [MySQL](https://www.mysql.com), [Gin](https://github.com/gin-gonic/gin), [ent.](https://entgo.io), [GraphQL](https://graphql.org), [Swagger 2.0](https://github.com/swaggo/gin-swagger)


## Project structure

```plaintext
├── cmd
│   └── stocms             # Main program entry
├── docs                   # Documentation
├── infra                  # Infrastructure configurations
├── internal               # Internal application logic
│   ├── config             # Configuration files
│   ├── data               # Data handling
│   │   ├── ent            # ent ORM related
│   │   ├── graphql        # GraphQL schemas
│   │   ├── repository     # Repositories for data access
│   │   ├── schema         # Database schemas
│   │   └── structs        # Data structures
│   ├── graphql            # GraphQL resolvers and types
│   ├── handler            # Request handlers
│   ├── helper             # Helper utilities and functions
│   ├── server             # Server-related code
│   │   └── middleware     # Middleware
│   └── service            # Business logic
└── pkg                    # Public packages
    ├── consts             # Constants
    ├── cookie             # Cookie handling
    ├── ecode              # Error codes
│   ├── email              # Email related logic
    ├── encrypt            # Encryption
    ├── jwt                # JWT handling
    ├── log                # Logging
    ├── nanoid             # NanoID generation
│   ├── oauth              # OAuth related logic
    ├── resp               # Response handling
    ├── slug               # Slug generation
    ├── time               # Time utilities
    ├── types              # Type definitions
    ├── util               # Utility functions
    ├── uuid               # UUID generation
    └── validator          # Validators
```


## Documentation

For full documentation, visit [https://domain.com](https://domain.com).

## Maintainers

[@Shen](https://github.com/haiyon)

## License

[MIT](LICENSE)

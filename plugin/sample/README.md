# Sample Plugin

> This plugin handles resource-related functionalities.

## Structure

```plaintext
├── cmd/                    # Command line related code
│   └── plugin.go           # Plugin instance symbol
├── data/                   # Data handling
│   ├── ent/                # ent ORM related
│   ├── repository/         # Repositories for data access
│   └── schema/             # Database schemas
├── handler/                # Request handlers
├── service/                # Business logic services
└── structs/                # Data structures

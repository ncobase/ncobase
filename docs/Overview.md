# Overview

## Project structure

```plaintext.
.
├── bin/                # Compiled binaries and executables
│   └── plugins/        # Plugin binary files
│
├── cmd/                # Command-line applications
│   ├── cli/            # CLI tools and utilities
│   └── ncobase/        # Main program entry point
│
├── core/               # Core system modules
│   ├── access/         # Access control and permissions
│   ├── auth/           # Authentication and authorization
│   ├── organization/   # Organization(Workspace) management
│   ├── system/         # System-wide functionalities
│   ├── space/          # Multi-space support
│   └── user/           # User management
│
├── docs/               # Documentation files
│   └── swagger/        # API documentation
│
├── domain/              # Business domain modules
│   ├── content/        # Content management
│   ├── realtime/       # Real-time functionality
│   └── resource/       # Resource management
│
├── front/              # Frontend codebase
│   ├── apps/           # Frontend applications
│   ├── packages/       # Shared frontend packages
│   └── plugins/        # Frontend plugins
│
├── logs/               # Application log files
│
├── pkg/                # Shared packages and utilities
│   ├── biz/            # Business logic utilities
│   ├── cache/          # Caching functionality
│   ├── config/         # Configuration utilities
│   ├── consts/         # Constants and enums
│   ├── cookie/         # Cookie management
│   ├── crypto/         # Cryptography utilities
│   ├── data/           # Data handling and processing
│   ├── ecode/          # Error codes and definitions
│   ├── elastic/        # Elasticsearch integration
│   ├── email/          # Email handling and delivery
│   ├── entgo/          # Ent ORM utilities
│   ├── extension/        # Extension and plugin interfaces and managers
│   ├── helper/         # Helper functions and utilities
│   ├── jwt/            # JWT authentication handling
│   ├── log/            # Logging utilities and interfaces
│   ├── meili/          # Meilisearch integration
│   ├── nanoid/         # Unique ID generation
│   ├── oauth/          # OAuth implementation
│   ├── observes/       # Observability tools
│   ├── paging/         # Pagination utilities
│   ├── proxy/          # Proxy interfaces and managers
│   ├── resp/           # Response handling utilities
│   ├── router/         # Routing utilities and middleware
│   ├── slug/           # URL slug generation
│   ├── storage/        # Storage management interfaces
│   ├── time/           # Time handling utilities
│   ├── types/          # Common type definitions
│   ├── util/           # General utility functions
│   ├── uuid/           # UUID generation utilities
│   └── validator/      # Data validation utilities
│
├── plugin/             # Plugin system
│   ├── counter/        # Counter plugin implementation
│   └── sample/         # Sample plugin template
│
├── proxy/              # API proxy functionality
│
└── setup/              # System setup and initialization
    ├── config/         # Configuration templates and samples
    ├── database/       # Database initialization resources
    ├── systemd/        # Systemd service configuration
    └── docker/         # Docker configuration files
```

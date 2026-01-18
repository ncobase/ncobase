# Overview

Ncobase is a modular Go application platform built on the ncore framework.

## Architecture

```text
ncobase/
├── core/           # Core modules (access, auth, organization, space, system, user)
├── biz/            # Business modules (content, realtime)
├── plugin/         # Plugins (counter, payment, proxy, resource, sample)
├── cmd/            # Application entry points
├── internal/       # Private code (middleware, server)
├── docs/           # Documentation
└── frontend/       # Web frontend
```

## Tech Stack

| Category       | Technologies                               |
| -------------- | ------------------------------------------ |
| Runtime        | Go 1.24+, Gin                              |
| Database       | PostgreSQL/MySQL (Ent ORM), MongoDB, Redis |
| Search         | Elasticsearch, OpenSearch, Meilisearch     |
| API            | REST, GraphQL, Swagger                     |
| Authentication | JWT, OAuth, Casbin RBAC, MFA               |
| Infrastructure | Docker, OpenTelemetry, Consul              |

## Modules

| Type     | Location  | Modules                                         |
| -------- | --------- | ----------------------------------------------- |
| Core     | `core/`   | access, auth, organization, space, system, user |
| Business | `biz/`    | content, realtime                               |
| Plugin   | `plugin/` | counter, payment, proxy, resource, sample       |

## Quick Start

```bash
make install && make generate
make run
```

## Documentation

- [Domain Reference](DOMAIN_REFERENCE.md) - API routing and module organization
- [Extension Guide](EXTENSION_GUIDE.md) - Module and plugin development
- [Migration Notes](MIGRATION_NOTES.md) - ncore v0.2 migration

# Ncobase Overview

Ncobase is a modular Go application platform built for enterprise-scale systems with microservices architecture.

## Architecture

```text
ncobase/
├── core/           # Core modules (access, auth, organization, space, system, user, etc...)
├── biz/            # Business modules (content, realtime, etc...)
├── plugin/         # Plugin system (counter, initialize, payment, proxy, resource, sample, etc...)
├── cmd/            # Application entry points
├── internal/       # Private application code (middleware, server, version)
├── docs/           # Documentation and API specs
├── examples/       # Example configurations and services
├── bin/            # Compiled binaries
└── test/           # Test files and utilities
```

## Key Technologies

- **Runtime**: Go 1.24+, Gin web framework
- **Database**: PostgreSQL/MySQL with Ent ORM
- **API**: REST + GraphQL with Swagger documentation
- **Authentication**: JWT, OAuth, Casbin RBAC
- **Infrastructure**: Docker, OpenTelemetry, Redis

## Module System

- **Core Modules**: Essential system functionality (access, auth, organization, space, system, user, etc...)
- **Business Modules**: Domain-specific logic (content, realtime, etc...)
- **Plugin System**: Extensible architecture (counter, initialize, payment, proxy, resource, sample, etc...)

## Quick Start

```bash
go mod tidy && go work sync
make install && make generate
make run
```

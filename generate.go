package ncobase

// Generate ent schema with versioned migrations
//go:generate go run entgo.io/ent/cmd/ent generate --feature sql/versioned-migration --target core/data/ent ncobase/core/data/schema

// Generate GraphQL server code
//go:generate go run github.com/99designs/gqlgen

// Generate Swagger documentation
//go:generate make swagger

// Generate components and plugins
//go:generate go generate ./feature/menu
//go:generate go generate ./feature/asset
//go:generate go generate ./feature/content

package ncobase

// Generate ent schema with versioned migrations
//go:generate go run entgo.io/ent/cmd/ent generate --feature sql/versioned-migration --target core/data/ent ncobase/core/data/schema

// Generate GraphQL server code
//go:generate go run github.com/99designs/gqlgen

// Generate Swagger documentation
//go:generate make swagger

// Generate plugins
//go:generate go generate ./plugin/asset
//go:generate go generate ./plugin/content

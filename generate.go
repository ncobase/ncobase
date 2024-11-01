package ncobase

// Generate GraphQL server code
// //go:generate go run github.com/99designs/gqlgen

// Generate features
//go:generate go generate ./core/access
//go:generate go generate ./core/auth
//go:generate go generate ./core/space
//go:generate go generate ./domain/realtime
//go:generate go generate ./core/system
//go:generate go generate ./core/tenant
//go:generate go generate ./core/user

// Generate business
//go:generate go generate ./domain/content
//go:generate go generate ./domain/resource

// Generate plugins
//go:generate go generate ./plugin/counter
//go:generate go generate ./plugin/sample

// Generate Swagger documentation
//go:generate make swagger

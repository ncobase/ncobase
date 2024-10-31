package ncobase

// Generate GraphQL server code
// //go:generate go run github.com/99designs/gqlgen

// Generate features
//go:generate go generate ./feature/access
//go:generate go generate ./feature/auth
//go:generate go generate ./feature/group
//go:generate go generate ./feature/socket
//go:generate go generate ./feature/system
//go:generate go generate ./feature/tenant
//go:generate go generate ./feature/user

// Generate business
//go:generate go generate ./domain/content
//go:generate go generate ./domain/resource

// Generate plugins
//go:generate go generate ./plugin/counter
//go:generate go generate ./plugin/sample

// Generate Swagger documentation
//go:generate make swagger

package ncobase

// Generate GraphQL server code
// //go:generate go run github.com/99designs/gqlgen

// Generate features
//go:generate go generate ./feature/access
//go:generate go generate ./feature/auth
//go:generate go generate ./feature/content
//go:generate go generate ./feature/group
//go:generate go generate ./feature/linker
//go:generate go generate ./feature/resource
//go:generate go generate ./feature/system
//go:generate go generate ./feature/tenant
//go:generate go generate ./feature/user

// Generate plugins
//go:generate go generate ./plugin/counter

// Generate Swagger documentation
//go:generate make swagger

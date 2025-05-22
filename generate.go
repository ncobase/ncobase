package ncobase

// Generate GraphQL server code
// //go:generate go run github.com/99designs/gqlgen

// Generate core components
//go:generate go generate ./core/access
//go:generate go generate ./core/auth
//go:generate go generate ./core/payment
//go:generate go generate ./core/realtime
//go:generate go generate ./core/space
//go:generate go generate ./core/system
//go:generate go generate ./core/tenant
//go:generate go generate ./core/user
//go:generate go generate ./core/workflow

// Generate business components
//go:generate go generate ./domain/content
//go:generate go generate ./domain/resource

// Generate plugins
//go:generate go generate ./plugin/counter
//go:generate go generate ./plugin/proxy

// Generate Swagger documentation
//go:generate make swagger

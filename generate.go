package ncobase

// Generate GraphQL server code
// //go:generate go run github.com/99designs/gqlgen

// Generate core components
//go:generate go generate ./core/access
//go:generate go generate ./core/auth
//go:generate go generate ./core/organization
//go:generate go generate ./core/system
//go:generate go generate ./core/space
//go:generate go generate ./core/user

// Generate business components
//go:generate go generate ./biz/content
//go:generate go generate ./biz/realtime

// Generate plugins
//go:generate go generate ./plugin/counter
//go:generate go generate ./plugin/payment
//go:generate go generate ./plugin/proxy
//go:generate go generate ./plugin/resource

// // Generate Swagger documentation
// //go:generate make swagger

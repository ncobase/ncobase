package bootstrap

import (
	"ncobase/app/graphql/generated"
	graph "ncobase/app/graphql/resolvers"
	"ncobase/app/service"
	"net/http"

	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// newGraphQLServer creates a GraphQL server.
func newGraphQLServer(es graphql.ExecutableSchema) *handler.Server {
	s := handler.New(es)
	s.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		KeepAlivePingInterval: 15 * time.Second,
	})
	s.AddTransport(transport.Options{})
	s.AddTransport(transport.GET{})
	s.AddTransport(transport.POST{})
	s.AddTransport(transport.MultipartForm{})
	s.SetQueryCache(lru.New(1000))
	s.Use(extension.Introspection{})
	s.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})

	return s
}

// graphqlHandler defines the GraphQL handler.
func graphqlHandler(svc *service.Service) gin.HandlerFunc {
	config := generated.Config{
		Resolvers: &graph.Resolver{
			Svc: svc,
		},
	}
	s := newGraphQLServer(generated.NewExecutableSchema(config))
	return func(c *gin.Context) {
		s.ServeHTTP(c.Writer, c.Request)
	}
}

// playgroundHandler defines the Playground handler.
func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL playground", "/graphql")
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// registerGraphql registers the GraphQL router.
func registerGraphql(e *gin.Engine, svc *service.Service, mode string) {
	g := e.Group("/graphql")
	g.POST("", graphqlHandler(svc))
	if !strings.Contains("release", mode) {
		g.GET("", playgroundHandler())
	}
}

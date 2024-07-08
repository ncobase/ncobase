// feature/interface.go
package feature

import (
	"ncobase/common/config"

	"github.com/gin-gonic/gin"
)

// Handler represents the handler for a feature
type Handler any

// Service represents the service for a feature
type Service any

// Interface defines the structure for a feature (Plugin / Module)
type Interface interface {
	Name() string
	Init(conf *config.Config) error
	RegisterRoutes(router *gin.Engine)
	GetHandlers() map[string]Handler
	GetServices() map[string]Service
	Cleanup() error
	Status() string
	GetMetadata() Metadata
}

// Metadata represents the metadata of a feature
type Metadata struct {
	Name         string   `json:"name,omitempty"`
	Version      string   `json:"version,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
	Description  string   `json:"description,omitempty"`
}

// Wrapper wraps a Interface instance with its metadata
type Wrapper struct {
	Metadata Metadata  `json:"metadata"`
	Instance Interface `json:"instance,omitempty"`
}

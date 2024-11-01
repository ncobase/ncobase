package templates

import "fmt"

func DataTemplate(name, moduleType string) string {
	return fmt.Sprintf(`package data

import (
	"context"
	"database/sql"
	"ncobase/common/config"
	"ncobase/common/data"
	"ncobase/common/elastic"
	"ncobase/common/meili"

	"github.com/redis/go-redis/v9"
)

// Data .
type Data struct {
	*data.Data
}

// New creates a new Database Connection.
func New(conf *config.Data) (*Data, func(name ...string), error) {
	d, cleanup, err := data.New(conf)
	if err != nil {
		return nil, nil, err
	}

	return &Data{
		Data: d,
	}, cleanup, nil
}

// Close closes all the resources in Data and returns any errors encountered.
func (d *Data) Close() (errs []error) {
	return errs
}

// GetDB get database
func (d *Data) GetDB() *sql.DB {
	return d.Conn.DB
}

// GetRedis get redis
func (d *Data) GetRedis() *redis.Client {
	return d.Conn.RC
}

// GetMeilisearch get meilisearch
func (d *Data) GetMeilisearch() *meili.Client {
	return d.Conn.MS
}

func (d *Data) GetElasticsearchClient() *elastic.Client {
	return d.Conn.ES
}

// Ping .
func (d *Data) Ping(ctx context.Context) error {
	return d.Conn.DB.PingContext(ctx)
}

// CloseDB .
func (d *Data) CloseDB() error {
	return d.Conn.DB.Close()
}
`)
}

func RepositoryTemplate(name, moduleType string) string {
	return fmt.Sprintf(`package repository

import "ncobase/%s/%s/data"

// Repository represents the %s repository.
type Repository struct {
	// Add your repository fields here
}

// New creates a new repository.
func New(d *data.Data) *Repository {
	return &Repository{
		// Initialize your repository fields here
	}
}

// Add your repository methods here
`, moduleType, name, name)
}

func SchemaTemplate() string {
	return `package schema

// Add your schema definitions here
`
}

func HandlerTemplate(name, moduleType string) string {
	return fmt.Sprintf(`package handler

import "ncobase/%s/%s/service"

// Handler represents the %s handler.
type Handler struct {
	// Add your handler fields here
}

// New creates a new handler.
func New(s *service.Service) *Handler {
	return &Handler{
		// Initialize your handler fields here
	}
}

// Add your handler methods here
`, moduleType, name, name)
}

func ServiceTemplate(name, moduleType string) string {
	return fmt.Sprintf(`package service

import (
	"ncobase/common/config"
	"ncobase/%s/%s/data"
)

// Service represents the %s service.
type Service struct {
	// Add your service fields here
}

// New creates a new service.
func New(conf *config.Config, d *data.Data) *Service {
	return &Service{
		// Initialize your service fields here
	}
}

// Add your service methods here
`, moduleType, name, name)
}

func StructsTemplate() string {
	return `package structs

// Define your structs here
`
}

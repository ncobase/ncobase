package meili

import (
	"context"
	"stocms/pkg/log"

	"github.com/meilisearch/meilisearch-go"
)

// Client Meilisearch client
type Client struct {
	client *meilisearch.Client
}

// NewMeilisearch new Meilisearch client
func NewMeilisearch(host, apiKey string) *Client {
	ms := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   host,
		APIKey: apiKey,
	})
	return &Client{client: ms}
}

// Search search from Meilisearch
func (c *Client) Search(ctx context.Context, index, query string, options *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
	resp, err := c.client.Index(index).Search(query, options)
	if err != nil {
		log.Errorf(ctx, "Meilisearch search error: %v", err)
		return nil, err
	}
	return resp, nil
}

// IndexDocuments index document to Meilisearch
func (c *Client) IndexDocuments(ctx context.Context, index string, document any, primaryKey ...string) error {
	_, err := c.client.Index(index).AddDocuments(document, primaryKey...)
	if err != nil {
		log.Errorf(ctx, "Meilisearch index document error: %v", err)
		return err
	}
	return nil
}

// UpdateDocuments update document to Meilisearch
func (c *Client) UpdateDocuments(ctx context.Context, index string, document any, documentID string) error {
	_, err := c.client.Index(index).UpdateDocuments(document, documentID)
	if err != nil {
		log.Errorf(ctx, "Meilisearch update document error: %v", err)
		return err
	}
	return nil
}

// DeleteDocuments delete document from Meilisearch
func (c *Client) DeleteDocuments(ctx context.Context, index, documentID string) error {
	_, err := c.client.Index(index).DeleteDocument(documentID)
	if err != nil {
		log.Errorf(ctx, "Meilisearch delete document error: %v", err)
		return err
	}
	return nil
}

// GetClient get Meilisearch client
func (c *Client) GetClient() *meilisearch.Client {
	return c.client
}

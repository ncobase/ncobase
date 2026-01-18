// Run: go generate ./...
//
// Generate ent schema with versioned migrations
//
//go:generate go run entgo.io/ent/cmd/ent generate --feature sql/versioned-migration,sql/execquery,sql/upsert ./data/schema --target ./data/ent
package sample

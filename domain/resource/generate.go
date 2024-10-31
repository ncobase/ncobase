package resource

// Generate ent schema with versioned migrations
//go:generate go run entgo.io/ent/cmd/ent generate --feature sql/versioned-migration --target data/ent ncobase/domain/resource/data/schema

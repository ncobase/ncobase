package tenant

// Generate ent schema with versioned migrations
//go:generate go run entgo.io/ent/cmd/ent generate --feature sql/versioned-migration --target data/ent ncobase/tenant/data/schema

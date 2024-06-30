package generated

//go:generate go run entgo.io/ent/cmd/ent generate --feature sql/versioned-migration --target app/data/ent ncobase/app/data/schema
//go:generate go run github.com/99designs/gqlgen
//go:generate make swagger

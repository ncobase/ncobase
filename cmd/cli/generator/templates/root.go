package templates

import "fmt"

// func ModuleTemplate(name, moduleType string) string {
// 	return fmt.Sprintf(`module ncobase/%s/%s
//
// go 1.23.0
//
// `, moduleType, name)
// }

func GeneraterTemplate(name, moduleType string) string {
	return fmt.Sprintf(`package %s

// Generate ent schema with versioned migrations
// To generate, Runs need to be deleted comment
// //go:generate go run entgo.io/ent/cmd/ent generate --feature sql/versioned-migration --target data/ent ncobase/%s/%s/data/schema

`, name, moduleType, name)
}

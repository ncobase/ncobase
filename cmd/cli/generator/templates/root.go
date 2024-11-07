package templates

import "fmt"

// func ExtTemplate(name, extType string) string {
// 	return fmt.Sprintf(`module ncobase/%s/%s
//
// go 1.23.0
//
// `, extType, name)
// }

func GeneraterTemplate(name, extType string) string {
	return fmt.Sprintf(`package %s

// Generate ent schema with versioned migrations
// To generate, Runs need to be deleted comment
// //go:generate go run entgo.io/ent/cmd/ent generate --feature sql/versioned-migration --target data/ent ncobase/%s/%s/data/schema

`, name, extType, name)
}

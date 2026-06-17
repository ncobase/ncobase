package service

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	menuData "ncobase/plugin/initialize/data"
)

func TestInitializationDataPermissionAlignment(t *testing.T) {
	loaders := map[string]DataLoader{
		"website":    &WebsiteDataLoader{},
		"company":    &CompanyDataLoader{},
		"enterprise": &EnterpriseDataLoader{},
	}

	for mode, loader := range loaders {
		t.Run(mode+"_menu_permissions", func(t *testing.T) {
			issues := menuData.ValidateMenuPermissions(permissionKeysForTest(loader))
			if len(issues) == 0 {
				return
			}

			slugs := make([]string, 0, len(issues))
			for slug := range issues {
				slugs = append(slugs, slug)
			}
			sort.Strings(slugs)

			var details []string
			for _, slug := range slugs {
				details = append(details, fmt.Sprintf("%s: %s", slug, strings.Join(issues[slug], "; ")))
			}
			t.Fatalf("menu permissions are not defined for %s: %s", mode, strings.Join(details, " | "))
		})

		t.Run(mode+"_role_permission_mapping", func(t *testing.T) {
			defined := make(map[string]struct{})
			for _, permission := range loader.GetPermissions() {
				defined[permission.Name] = struct{}{}
			}

			var missing []string
			for role, permissions := range loader.GetRolePermissionMapping() {
				for _, permissionName := range permissions {
					if _, ok := defined[permissionName]; !ok {
						missing = append(missing, fmt.Sprintf("%s -> %s", role, permissionName))
					}
				}
			}

			if len(missing) > 0 {
				sort.Strings(missing)
				t.Fatalf("role permission mappings reference undefined permissions for %s: %s", mode, strings.Join(missing, ", "))
			}
		})
	}
}

func permissionKeysForTest(loader DataLoader) []string {
	return permissionKeysForDefinitions(loader.GetPermissions())
}

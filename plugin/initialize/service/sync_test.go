package service

import (
	"testing"

	accessStructs "ncobase/core/access/structs"
	systemStructs "ncobase/core/system/structs"
)

func TestDefaultPermissionUpdatesRepairsContractFields(t *testing.T) {
	defaultFalse := false
	defaultTrue := true
	disabledTrue := true
	disabledFalse := false

	updates := defaultPermissionUpdates(&accessStructs.ReadPermission{
		ID:          "perm-1",
		Name:        "Role Read",
		Action:      "manage",
		Subject:     "roles",
		Description: "Old description",
		Default:     &defaultFalse,
		Disabled:    &disabledTrue,
	}, accessStructs.CreatePermissionBody{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Role Read",
			Action:      "read",
			Subject:     "role",
			Description: "View roles, role permissions, and role user assignments",
			Default:     &defaultTrue,
			Disabled:    &disabledFalse,
		},
	})

	if updates["action"] != "read" {
		t.Fatalf("expected action update, got %#v", updates["action"])
	}
	if updates["subject"] != "role" {
		t.Fatalf("expected subject update, got %#v", updates["subject"])
	}
	if updates["description"] == nil {
		t.Fatal("expected description update")
	}
	if updates["default"] != true {
		t.Fatalf("expected default=true update, got %#v", updates["default"])
	}
	if updates["disabled"] != false {
		t.Fatalf("expected disabled=false update, got %#v", updates["disabled"])
	}
}

func TestDefaultPermissionUpdatesNoopsWhenDefinitionsMatch(t *testing.T) {
	defaultTrue := true
	disabledFalse := false

	updates := defaultPermissionUpdates(&accessStructs.ReadPermission{
		ID:          "perm-1",
		Name:        "Role Read",
		Action:      "read",
		Subject:     "role",
		Description: "View roles",
		Default:     &defaultTrue,
		Disabled:    &disabledFalse,
	}, accessStructs.CreatePermissionBody{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Role Read",
			Action:      "read",
			Subject:     "role",
			Description: "View roles",
			Default:     &defaultTrue,
			Disabled:    &disabledFalse,
		},
	})

	if len(updates) != 0 {
		t.Fatalf("expected no updates, got %#v", updates)
	}
}

func TestMenuSyncIndexMatchesFallbackMenusWithoutSlug(t *testing.T) {
	parent := &systemStructs.ReadMenu{
		ID:   "menu-content",
		Slug: "content",
		Type: "header",
	}
	child := &systemStructs.ReadMenu{
		ID:       "menu-taxonomies",
		Name:     "Taxonomy",
		Label:    "content.taxonomies.navigation",
		Type:     "sidebar",
		Path:     "/content/taxonomies",
		Order:    98,
		ParentID: "menu-content",
	}

	index := newMenuSyncIndex([]*systemStructs.ReadMenu{parent, child})
	key := defaultMenuDefinitionKey(systemStructs.MenuBody{
		Name:     "Taxonomy",
		Label:    "content.taxonomies.navigation",
		Type:     "sidebar",
		Path:     "/content/taxonomies",
		Order:    intPtr(98),
		ParentID: "content",
	})

	if index.byKey[key] == nil || index.byKey[key].ID != child.ID {
		t.Fatalf("expected fallback key to resolve existing child menu, got %#v", index.byKey[key])
	}
}

func TestDefaultMenuUpdatesRepairsPermissionAndParent(t *testing.T) {
	hiddenFalse := false
	disabledFalse := false
	order := 970

	updates, changed := defaultMenuUpdates(&systemStructs.ReadMenu{
		ID:       "system-roles",
		Name:     "Role",
		Label:    "system.roles.navigation",
		Slug:     "system-roles",
		Type:     "sidebar",
		Path:     "/system/roles",
		Icon:     "IconUsersGroup",
		Perms:    "manage:roles",
		Order:    900,
		ParentID: "old-parent",
	}, systemStructs.MenuBody{
		Name:     "Role",
		Label:    "system.roles.navigation",
		Slug:     "system-roles",
		Type:     "sidebar",
		Path:     "/system/roles",
		Icon:     "IconUsersGroup",
		Perms:    "read:roles",
		Order:    &order,
		Hidden:   &hiddenFalse,
		Disabled: &disabledFalse,
		ParentID: "system-parent",
	}, "admin-user")

	if !changed {
		t.Fatal("expected menu update to be detected")
	}
	if updates.Perms != "read:roles" {
		t.Fatalf("expected permission repair, got %q", updates.Perms)
	}
	if updates.ParentID != "system-parent" {
		t.Fatalf("expected parent repair, got %q", updates.ParentID)
	}
	if updates.Order == nil || *updates.Order != 970 {
		t.Fatalf("expected order repair, got %#v", updates.Order)
	}
	if updates.UpdatedBy == nil || *updates.UpdatedBy != "admin-user" {
		t.Fatalf("expected updated_by to be assigned, got %#v", updates.UpdatedBy)
	}
}

func intPtr(value int) *int {
	return &value
}

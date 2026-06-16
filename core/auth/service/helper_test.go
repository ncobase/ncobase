package service

import (
	"context"
	"sort"
	"testing"

	accessStructs "ncobase/core/access/structs"
	spaceStructs "ncobase/core/space/structs"

	"github.com/ncobase/ncore/ctxutil"
)

type fakeAuthAccessProvider struct {
	globalRoles []*accessStructs.ReadRole
	rolesByID   map[string]*accessStructs.ReadRole
	permissions map[string][]*accessStructs.ReadPermission
}

func (f *fakeAuthAccessProvider) GetUserRoles(ctx context.Context, userID string) ([]*accessStructs.ReadRole, error) {
	return f.globalRoles, nil
}

func (f *fakeAuthAccessProvider) GetByIDs(ctx context.Context, roleIDs []string) ([]*accessStructs.ReadRole, error) {
	roles := make([]*accessStructs.ReadRole, 0, len(roleIDs))
	for _, id := range roleIDs {
		if role := f.rolesByID[id]; role != nil {
			roles = append(roles, role)
		}
	}
	return roles, nil
}

func (f *fakeAuthAccessProvider) GetRolePermissions(ctx context.Context, roleID string) ([]*accessStructs.ReadPermission, error) {
	return f.permissions[roleID], nil
}

type fakeAuthSpaceProvider struct {
	defaultSpace *spaceStructs.ReadSpace
	roleIDs      map[string][]string
}

func (f *fakeAuthSpaceProvider) GetUserSpace(ctx context.Context, userID string) (*spaceStructs.ReadSpace, error) {
	return f.defaultSpace, nil
}

func (f *fakeAuthSpaceProvider) GetUserRolesInSpace(ctx context.Context, userID, spaceID string) ([]string, error) {
	return f.roleIDs[spaceID], nil
}

func TestGetUserSpacesRolesPermissionsIncludesSpaceRolePermissions(t *testing.T) {
	ctx := ctxutil.SetSpaceID(context.Background(), "space-b")
	accessProvider := &fakeAuthAccessProvider{
		globalRoles: []*accessStructs.ReadRole{
			{ID: "role-global", Slug: "reader"},
		},
		rolesByID: map[string]*accessStructs.ReadRole{
			"role-space": {ID: "role-space", Slug: "space-manager"},
		},
		permissions: map[string][]*accessStructs.ReadPermission{
			"role-global": {
				{ID: "perm-read-users", Action: "read", Subject: "users"},
			},
			"role-space": {
				{ID: "perm-manage-spaces", Action: "manage", Subject: "spaces"},
			},
		},
	}
	spaceProvider := &fakeAuthSpaceProvider{
		defaultSpace: &spaceStructs.ReadSpace{ID: "space-a", Name: "A"},
		roleIDs: map[string][]string{
			"space-b": {"role-space"},
		},
	}

	spaceID, roles, permissions, isAdmin, err := getUserSpacesRolesPermissions(
		ctx,
		"user-1",
		accessProvider,
		spaceProvider,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if spaceID != "space-b" {
		t.Fatalf("expected requested space to be retained, got %q", spaceID)
	}
	if isAdmin {
		t.Fatal("expected non-admin role set")
	}

	assertStringContains(t, roles, "reader")
	assertStringContains(t, roles, "space-manager")
	assertStringContains(t, permissions, "read:users")
	assertStringContains(t, permissions, "manage:spaces")
}

func TestGetUserSpacesRolesPermissionsFallsBackToDefaultSpace(t *testing.T) {
	accessProvider := &fakeAuthAccessProvider{
		rolesByID: map[string]*accessStructs.ReadRole{
			"role-default": {ID: "role-default", Slug: "default-space-reader"},
		},
		permissions: map[string][]*accessStructs.ReadPermission{
			"role-default": {
				{ID: "perm-read-spaces", Action: "read", Subject: "spaces"},
			},
		},
	}
	spaceProvider := &fakeAuthSpaceProvider{
		defaultSpace: &spaceStructs.ReadSpace{ID: "space-default", Name: "Default"},
		roleIDs: map[string][]string{
			"space-default": {"role-default"},
		},
	}

	spaceID, roles, permissions, _, err := getUserSpacesRolesPermissions(
		context.Background(),
		"user-1",
		accessProvider,
		spaceProvider,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if spaceID != "space-default" {
		t.Fatalf("expected default space fallback, got %q", spaceID)
	}
	assertStringContains(t, roles, "default-space-reader")
	assertStringContains(t, permissions, "read:spaces")
}

func TestFindUserSpace(t *testing.T) {
	spaces := []*spaceStructs.ReadSpace{
		{ID: "space-a", Name: "A"},
		{ID: "space-b", Name: "B"},
	}

	if got := findUserSpace(spaces, "space-b"); got == nil || got.Name != "B" {
		t.Fatalf("expected to find space-b, got %#v", got)
	}
	if got := findUserSpace(spaces, "space-c"); got != nil {
		t.Fatalf("expected missing space to return nil, got %#v", got)
	}
}

func TestResolveActiveSpaceKeepsRequestedUserSpace(t *testing.T) {
	spaces := []*spaceStructs.ReadSpace{
		{ID: "space-a", Name: "A"},
		{ID: "space-b", Name: "B"},
	}
	spaceProvider := &fakeAuthSpaceProvider{
		defaultSpace: &spaceStructs.ReadSpace{ID: "space-a", Name: "A"},
	}

	got := resolveActiveSpace(ctxutil.SetSpaceID(context.Background(), "space-b"), "user-1", spaces, spaceProvider)
	if got == nil || got.ID != "space-b" {
		t.Fatalf("expected requested user space to be retained, got %#v", got)
	}
}

func TestResolveActiveSpaceFallsBackWhenRequestedSpaceIsNotOwned(t *testing.T) {
	spaces := []*spaceStructs.ReadSpace{
		{ID: "space-a", Name: "A"},
		{ID: "space-b", Name: "B"},
	}
	spaceProvider := &fakeAuthSpaceProvider{
		defaultSpace: &spaceStructs.ReadSpace{ID: "space-a", Name: "A"},
	}

	got := resolveActiveSpace(ctxutil.SetSpaceID(context.Background(), "space-c"), "user-1", spaces, spaceProvider)
	if got == nil || got.ID != "space-a" {
		t.Fatalf("expected default space fallback, got %#v", got)
	}
}

func TestResolveActiveSpaceFallsBackToFirstUserSpaceWithoutDefault(t *testing.T) {
	spaces := []*spaceStructs.ReadSpace{
		{ID: "space-a", Name: "A"},
		{ID: "space-b", Name: "B"},
	}

	got := resolveActiveSpace(context.Background(), "user-1", spaces, nil)
	if got == nil || got.ID != "space-a" {
		t.Fatalf("expected first user space fallback, got %#v", got)
	}
}

func assertStringContains(t *testing.T, values []string, expected string) {
	t.Helper()
	for _, value := range values {
		if value == expected {
			return
		}
	}
	sort.Strings(values)
	t.Fatalf("expected %q in %v", expected, values)
}

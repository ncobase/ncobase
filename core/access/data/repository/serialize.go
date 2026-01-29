package repository

import (
	"ncobase/core/access/data/ent"
	"ncobase/core/access/structs"
)

// SerializeRole converts ent.Role to structs.ReadRole.
func SerializeRole(row *ent.Role) *structs.ReadRole {
	if row == nil {
		return nil
	}
	return &structs.ReadRole{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Disabled:    row.Disabled,
		Description: row.Description,
		Extras:      &row.Extras,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}
}

// SerializeRoles converts []*ent.Role to []*structs.ReadRole.
func SerializeRoles(rows []*ent.Role) []*structs.ReadRole {
	rs := make([]*structs.ReadRole, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, SerializeRole(row))
	}
	return rs
}

// SerializePermission converts ent.Permission to structs.ReadPermission.
func SerializePermission(row *ent.Permission) *structs.ReadPermission {
	if row == nil {
		return nil
	}
	return &structs.ReadPermission{
		ID:          row.ID,
		Name:        row.Name,
		Action:      row.Action,
		Subject:     row.Subject,
		Description: row.Description,
		Default:     &row.Default,
		Disabled:    &row.Disabled,
		Extras:      &row.Extras,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}
}

// SerializePermissions converts []*ent.Permission to []*structs.ReadPermission.
func SerializePermissions(rows []*ent.Permission) []*structs.ReadPermission {
	rs := make([]*structs.ReadPermission, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, SerializePermission(row))
	}
	return rs
}

// SerializeCasbinRule converts ent.CasbinRule to structs.ReadCasbinRule.
func SerializeCasbinRule(row *ent.CasbinRule) *structs.ReadCasbinRule {
	if row == nil {
		return nil
	}
	return &structs.ReadCasbinRule{
		PType:     row.PType,
		V0:        row.V0,
		V1:        row.V1,
		V2:        row.V2,
		V3:        &row.V3,
		V4:        &row.V4,
		V5:        &row.V5,
		CreatedBy: &row.CreatedBy,
		CreatedAt: &row.CreatedAt,
		UpdatedBy: &row.UpdatedBy,
		UpdatedAt: &row.UpdatedAt,
	}
}

// SerializeCasbinRules converts []*ent.CasbinRule to []*structs.ReadCasbinRule.
func SerializeCasbinRules(rows []*ent.CasbinRule) []*structs.ReadCasbinRule {
	rs := make([]*structs.ReadCasbinRule, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, SerializeCasbinRule(row))
	}
	return rs
}

// SerializeRolePermission converts ent.RolePermission to structs.RolePermission.
func SerializeRolePermission(row *ent.RolePermission) *structs.RolePermission {
	if row == nil {
		return nil
	}
	return &structs.RolePermission{
		RoleID:       row.RoleID,
		PermissionID: row.PermissionID,
	}
}

// SerializeRolePermissions converts []*ent.RolePermission to []*structs.RolePermission.
func SerializeRolePermissions(rows []*ent.RolePermission) []*structs.RolePermission {
	rs := make([]*structs.RolePermission, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, SerializeRolePermission(row))
	}
	return rs
}

// SerializeUserRole converts ent.UserRole to structs.UserRole.
func SerializeUserRole(row *ent.UserRole) *structs.UserRole {
	if row == nil {
		return nil
	}
	return &structs.UserRole{
		UserID: row.UserID,
		RoleID: row.RoleID,
	}
}

// SerializeUserRoles converts []*ent.UserRole to []*structs.UserRole.
func SerializeUserRoles(rows []*ent.UserRole) []*structs.UserRole {
	rs := make([]*structs.UserRole, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, SerializeUserRole(row))
	}
	return rs
}

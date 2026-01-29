package repository

import (
	"ncobase/core/organization/data/ent"
	"ncobase/core/organization/structs"
)

// SerializeOrganization converts ent.Organization to structs.ReadOrganization.
func SerializeOrganization(row *ent.Organization) *structs.ReadOrganization {
	if row == nil {
		return nil
	}
	return &structs.ReadOrganization{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Type:        row.Type,
		Disabled:    row.Disabled,
		Description: row.Description,
		Leader:      &row.Leader,
		Extras:      &row.Extras,
		ParentID:    &row.ParentID,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}
}

// SerializeOrganizations converts []*ent.Organization to []*structs.ReadOrganization.
func SerializeOrganizations(rows []*ent.Organization) []*structs.ReadOrganization {
	rs := make([]*structs.ReadOrganization, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, SerializeOrganization(row))
	}
	return rs
}

// SerializeOrganizationRole converts ent.OrganizationRole to structs.OrganizationRole.
func SerializeOrganizationRole(row *ent.OrganizationRole) *structs.OrganizationRole {
	if row == nil {
		return nil
	}
	return &structs.OrganizationRole{
		OrgID:  row.OrgID,
		RoleID: row.RoleID,
	}
}

// SerializeOrganizationRoles converts []*ent.OrganizationRole to []*structs.OrganizationRole.
func SerializeOrganizationRoles(rows []*ent.OrganizationRole) []*structs.OrganizationRole {
	rs := make([]*structs.OrganizationRole, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, SerializeOrganizationRole(row))
	}
	return rs
}

// SerializeUserOrganization converts ent.UserOrganization to structs.UserOrganization.
func SerializeUserOrganization(row *ent.UserOrganization) *structs.UserOrganization {
	if row == nil {
		return nil
	}
	return &structs.UserOrganization{
		UserID: row.UserID,
		OrgID:  row.OrgID,
	}
}

// SerializeUserOrganizations converts []*ent.UserOrganization to []*structs.UserOrganization.
func SerializeUserOrganizations(rows []*ent.UserOrganization) []*structs.UserOrganization {
	rs := make([]*structs.UserOrganization, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, SerializeUserOrganization(row))
	}
	return rs
}

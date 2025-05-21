package data

import accessStructs "ncobase/access/structs"

// SystemDefaultRoles defines the system default roles
var SystemDefaultRoles = []accessStructs.CreateRoleBody{
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Super Administrator",
			Slug:        "super-admin",
			Disabled:    false,
			Description: "Full system access with all permissions",
			Extras:      nil,
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "System Administrator",
			Slug:        "admin",
			Disabled:    false,
			Description: "General administrative access to the system",
			Extras:      nil,
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Standard User",
			Slug:        "user",
			Disabled:    false,
			Description: "Regular user with basic permissions",
			Extras:      nil,
		},
	},
}

// OrganizationRoles defines organization-specific roles
var OrganizationRoles = []OrganizationRole{
	{
		Role: accessStructs.RoleBody{
			Name:        "Organization Admin",
			Slug:        "org-admin",
			Disabled:    false,
			Description: "Full access to organization management",
			Extras:      nil,
		},
		Permissions: []string{
			"Manage Organization",
			"Manage Members",
			"View Organization",
		},
	},
	{
		Role: accessStructs.RoleBody{
			Name:        "Team Admin",
			Slug:        "team-admin",
			Disabled:    false,
			Description: "Admin for team-level organizations",
			Extras:      nil,
		},
		Permissions: []string{
			"Manage Organization",
			"Manage Members",
			"View Organization",
		},
	},
	{
		Role: accessStructs.RoleBody{
			Name:        "Member",
			Slug:        "member",
			Disabled:    false,
			Description: "Regular member of an organization",
			Extras:      nil,
		},
		Permissions: []string{
			"View Organization",
		},
	},
}

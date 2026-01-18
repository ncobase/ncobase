package website

import accessStructs "ncobase/core/access/structs"

// SystemDefaultRoles for regular websites
var SystemDefaultRoles = []accessStructs.CreateRoleBody{
	// System level
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Super Administrator",
			Slug:        "super-admin",
			Description: "System super admin with full access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Administrator",
			Slug:        "admin",
			Description: "Site administrator with management access",
		},
	},

	// Website level
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Manager",
			Slug:        "manager",
			Description: "Content and user manager",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Member",
			Slug:        "member",
			Description: "Registered member with standard access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Viewer",
			Slug:        "viewer",
			Description: "Read-only access user",
		},
	},
}

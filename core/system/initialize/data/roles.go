package data

import accessStructs "ncobase/core/access/structs"

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

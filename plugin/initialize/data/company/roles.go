package company

import accessStructs "ncobase/access/structs"

// SystemDefaultRoles defines core system roles
var SystemDefaultRoles = []accessStructs.CreateRoleBody{
	// System level roles
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Super Administrator",
			Slug:        "super-admin",
			Description: "System super admin with unrestricted access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "System Administrator",
			Slug:        "system-admin",
			Description: "System admin with platform management access",
		},
	},

	// Enterprise level roles
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Enterprise Admin",
			Slug:        "enterprise-admin",
			Description: "Enterprise admin with business management access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Manager",
			Slug:        "manager",
			Description: "Manager with team and department oversight",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Employee",
			Slug:        "employee",
			Description: "Standard employee with basic platform access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Guest",
			Slug:        "guest",
			Description: "Guest user with read-only access",
		},
	},
}

package enterprise

import accessStructs "ncobase/access/structs"

// SystemDefaultRoles defines simplified enterprise system roles
var SystemDefaultRoles = []accessStructs.CreateRoleBody{
	// System level roles
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Super Administrator",
			Slug:        "super-admin",
			Disabled:    false,
			Description: "Super administrator with unrestricted system access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "System Administrator",
			Slug:        "system-admin",
			Disabled:    false,
			Description: "System administrator with full platform management",
		},
	},

	// Enterprise level roles
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Enterprise Administrator",
			Slug:        "enterprise-admin",
			Disabled:    false,
			Description: "Enterprise-wide administrative access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Department Manager",
			Slug:        "department-manager",
			Disabled:    false,
			Description: "Department-level management and oversight",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Team Leader",
			Slug:        "team-leader",
			Disabled:    false,
			Description: "Team leadership and coordination",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Employee",
			Slug:        "employee",
			Disabled:    false,
			Description: "Standard employee with basic platform access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Contractor",
			Slug:        "contractor",
			Disabled:    false,
			Description: "External contractor with limited access",
		},
	},
}

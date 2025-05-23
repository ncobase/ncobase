package data

import accessStructs "ncobase/access/structs"

// SystemDefaultRoles defines enterprise system roles
var SystemDefaultRoles = []accessStructs.CreateRoleBody{
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Super Administrator",
			Slug:        "super-admin",
			Disabled:    false,
			Description: "Super administrator with full system access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "System Administrator",
			Slug:        "system-admin",
			Disabled:    false,
			Description: "Full system access and management",
		},
	},
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
			Name:        "Tenant Administrator",
			Slug:        "tenant-admin",
			Disabled:    false,
			Description: "Tenant-level administrative access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "HR Manager",
			Slug:        "hr-manager",
			Disabled:    false,
			Description: "Human resources management access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Finance Manager",
			Slug:        "finance-manager",
			Disabled:    false,
			Description: "Financial management access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Department Manager",
			Slug:        "department-manager",
			Disabled:    false,
			Description: "Department-level management access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Team Leader",
			Slug:        "team-leader",
			Disabled:    false,
			Description: "Team leadership access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Employee",
			Slug:        "employee",
			Disabled:    false,
			Description: "Standard employee access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Contractor",
			Slug:        "contractor",
			Disabled:    false,
			Description: "External contractor access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Intern",
			Slug:        "intern",
			Disabled:    false,
			Description: "Intern-level access",
		},
	},
}

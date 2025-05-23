package data

import accessStructs "ncobase/access/structs"

// SystemDefaultPermissions defines enterprise system permissions
var SystemDefaultPermissions = []accessStructs.CreatePermissionBody{
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Super Admin Access",
			Action:      "*",
			Subject:     "*",
			Description: "Super administrator wildcard permission",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "System Management",
			Action:      "manage",
			Subject:     "system",
			Description: "Manage system settings and configuration",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Tenant Management",
			Action:      "manage",
			Subject:     "tenant",
			Description: "Create, update, and delete tenants",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "User Management",
			Action:      "manage",
			Subject:     "user",
			Description: "Create, update, and delete users",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "User Read",
			Action:      "read",
			Subject:     "user",
			Description: "View user information",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "User Create",
			Action:      "create",
			Subject:     "user",
			Description: "Create new users",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "User Update",
			Action:      "update",
			Subject:     "user",
			Description: "Update user information",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "User Delete",
			Action:      "delete",
			Subject:     "user",
			Description: "Delete users",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Employee Management",
			Action:      "manage",
			Subject:     "employee",
			Description: "Manage employee records",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Organization Management",
			Action:      "manage",
			Subject:     "organization",
			Description: "Manage organizational structure",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Role Management",
			Action:      "manage",
			Subject:     "role",
			Description: "Create and manage roles",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Permission Management",
			Action:      "manage",
			Subject:     "permission",
			Description: "Manage permissions and access control",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Financial Management",
			Action:      "manage",
			Subject:     "finance",
			Description: "Manage financial records and transactions",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "HR Management",
			Action:      "manage",
			Subject:     "hr",
			Description: "Human resources management",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Dashboard Access",
			Action:      "read",
			Subject:     "dashboard",
			Description: "Access to dashboard and analytics",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Enterprise Overview",
			Action:      "read",
			Subject:     "enterprise",
			Description: "View enterprise-wide information",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Cross Company Access",
			Action:      "read",
			Subject:     "cross-company",
			Description: "Access information across multiple companies",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Group Management",
			Action:      "manage",
			Subject:     "group",
			Description: "Full group management access",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Group Read",
			Action:      "read",
			Subject:     "group",
			Description: "View group information",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Department Management",
			Action:      "manage",
			Subject:     "department",
			Description: "Full department management access",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Department Read",
			Action:      "read",
			Subject:     "department",
			Description: "View department information",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Team Management",
			Action:      "manage",
			Subject:     "team",
			Description: "Full team management access",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Team Read",
			Action:      "read",
			Subject:     "team",
			Description: "View team information",
		},
	},
}

// RolePermissionMapping defines role-permission relationships
var RolePermissionMapping = map[string][]string{
	"super-admin": {
		"Super Admin Access",
	},
	"system-admin": {
		"System Management",
		"Tenant Management",
		"User Management",
		"Employee Management",
		"Organization Management",
		"Role Management",
		"Permission Management",
		"Dashboard Access",
		"Group Management",
		"Department Management",
		"Team Management",
	},
	"enterprise-admin": {
		"User Management",
		"Employee Management",
		"Organization Management",
		"Financial Management",
		"HR Management",
		"Dashboard Access",
		"Enterprise Overview",
		"Cross Company Access",
		"Group Management",
		"Department Management",
		"Team Management",
	},
	"tenant-admin": {
		"User Read",
		"User Create",
		"User Update",
		"Employee Management",
		"Dashboard Access",
		"Group Read",
		"Department Read",
		"Team Read",
	},
	"hr-manager": {
		"Employee Management",
		"HR Management",
		"User Read",
		"Dashboard Access",
		"Department Read",
		"Team Read",
	},
	"finance-manager": {
		"Financial Management",
		"Employee Management",
		"Dashboard Access",
		"Department Read",
		"Team Read",
	},
	"department-manager": {
		"User Read",
		"Employee Management",
		"Dashboard Access",
		"Department Management",
		"Team Management",
	},
	"team-leader": {
		"User Read",
		"Dashboard Access",
		"Team Read",
	},
	"employee": {
		"Dashboard Access",
	},
	"contractor": {
		"Dashboard Access",
	},
	"intern": {
		"Dashboard Access",
	},
}

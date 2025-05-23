package company

import accessStructs "ncobase/access/structs"

// SystemDefaultPermissions defines core system permissions
var SystemDefaultPermissions = []accessStructs.CreatePermissionBody{
	// Super admin permission
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Super Admin Access",
			Action:      "*",
			Subject:     "*",
			Description: "Super admin wildcard permission",
		},
	},

	// Basic access permissions
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Dashboard Access",
			Action:      "read",
			Subject:     "dashboard",
			Description: "Access to dashboard",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Profile Management",
			Action:      "manage",
			Subject:     "profile",
			Description: "Manage own profile",
		},
	},

	// System management permissions
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "System Management",
			Action:      "manage",
			Subject:     "system",
			Description: "System configuration management",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Menu Management",
			Action:      "manage",
			Subject:     "menu",
			Description: "Menu structure management",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Dictionary Management",
			Action:      "manage",
			Subject:     "dictionary",
			Description: "Dictionary data management",
		},
	},

	// User management permissions
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "User Management",
			Action:      "manage",
			Subject:     "user",
			Description: "Full user management",
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
			Name:        "Employee Management",
			Action:      "manage",
			Subject:     "employee",
			Description: "Full employee management",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Employee Read",
			Action:      "read",
			Subject:     "employee",
			Description: "View employee information",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Employee Update",
			Action:      "update",
			Subject:     "employee",
			Description: "Update employee information",
		},
	},

	// Role and permission management
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Role Management",
			Action:      "manage",
			Subject:     "role",
			Description: "Role management",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Permission Read",
			Action:      "read",
			Subject:     "permission",
			Description: "View permissions",
		},
	},

	// Tenant management
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Tenant Management",
			Action:      "manage",
			Subject:     "tenant",
			Description: "Tenant management",
		},
	},

	// Organization management
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Organization Management",
			Action:      "manage",
			Subject:     "organization",
			Description: "Organization structure management",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Group Management",
			Action:      "manage",
			Subject:     "group",
			Description: "Group management",
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

	// Content management
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Content Management",
			Action:      "manage",
			Subject:     "content",
			Description: "Content management",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Content Read",
			Action:      "read",
			Subject:     "content",
			Description: "View content",
		},
	},

	// Workflow management
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Workflow Management",
			Action:      "manage",
			Subject:     "workflow",
			Description: "Workflow management",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Workflow Read",
			Action:      "read",
			Subject:     "workflow",
			Description: "View workflow",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Task Management",
			Action:      "manage",
			Subject:     "task",
			Description: "Task management",
		},
	},
}

// RolePermissionMapping defines role-permission relationships
var RolePermissionMapping = map[string][]string{
	"super-admin": {
		"Super Admin Access", // Wildcard permission
	},
	"system-admin": {
		"System Management",
		"Menu Management",
		"Dictionary Management",
		"User Management",
		"Employee Management",
		"Role Management",
		"Permission Read",
		"Tenant Management",
		"Organization Management",
		"Group Management",
		"Content Management",
		"Workflow Management",
		"Task Management",
		"Dashboard Access",
		"Profile Management",
	},
	"enterprise-admin": {
		"User Management",
		"Employee Management",
		"Organization Management",
		"Group Management",
		"Content Management",
		"Workflow Management",
		"Task Management",
		"Dashboard Access",
		"Profile Management",
		"Permission Read",
	},
	"manager": {
		"User Read",
		"Employee Read",
		"Employee Update",
		"Group Read",
		"Content Read",
		"Workflow Read",
		"Task Management",
		"Dashboard Access",
		"Profile Management",
	},
	"employee": {
		"User Read",
		"Employee Read",
		"Group Read",
		"Content Read",
		"Workflow Read",
		"Dashboard Access",
		"Profile Management",
	},
	"guest": {
		"Dashboard Access",
		"Content Read",
		"Profile Management",
	},
}

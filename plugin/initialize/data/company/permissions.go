package company

import accessStructs "ncobase/access/structs"

// SystemDefaultPermissions defines system permissions
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

	// Basic access permissions (all users need these)
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
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Account Management",
			Action:      "manage",
			Subject:     "account",
			Description: "Manage account settings",
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
			Name:        "System Read",
			Action:      "read",
			Subject:     "system",
			Description: "View system information",
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
			Name:        "Dictionary Read",
			Action:      "read",
			Subject:     "dictionary",
			Description: "View dictionary data",
		},
	},

	// User and Employee management
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
			Name:        "User Create",
			Action:      "create",
			Subject:     "user",
			Description: "Create new users",
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
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Employee Create",
			Action:      "create",
			Subject:     "employee",
			Description: "Create employee records",
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
			Name:        "Permission Management",
			Action:      "manage",
			Subject:     "permission",
			Description: "Permission management",
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
			Name:        "Organization Read",
			Action:      "read",
			Subject:     "organization",
			Description: "View organization information",
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
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Comment Read",
			Action:      "read",
			Subject:     "comment",
			Description: "View comments",
		},
	},

	// Module-specific permissions aligned with routes
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "TBP Management",
			Action:      "manage",
			Subject:     "tbp",
			Description: "TBP endpoints and routes management",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Realtime Access",
			Action:      "read",
			Subject:     "realtime",
			Description: "Access realtime features",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Resource Management",
			Action:      "manage",
			Subject:     "resource",
			Description: "Resource management",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Resource Read",
			Action:      "read",
			Subject:     "resource",
			Description: "View resources",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Payment Read",
			Action:      "read",
			Subject:     "payment",
			Description: "View payment information",
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
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "CMS Read",
			Action:      "read",
			Subject:     "cms",
			Description: "View CMS content",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "CMS Management",
			Action:      "manage",
			Subject:     "cms",
			Description: "CMS management",
		},
	},
}

// RolePermissionMapping aligned with company roles and actual menu permissions
var RolePermissionMapping = map[string][]string{
	"super-admin": {
		"Super Admin Access",
	},
	"system-admin": {
		// Full system access
		"System Management",
		"System Read",
		"Menu Management",
		"Dictionary Read",
		"User Management",
		"Employee Management",
		"Role Management",
		"Permission Management",
		"Permission Read",
		"Tenant Management",
		"Organization Read",
		"Group Management",
		"Group Read",
		"Content Management",
		"Content Read",
		"Comment Read",
		"TBP Management",
		"Realtime Access",
		"Resource Management",
		"Resource Read",
		"Payment Read",
		"Workflow Read",
		"Task Management",
		"CMS Management",
		"CMS Read",
		"Dashboard Access",
		"Profile Management",
		"Account Management",
	},
	"company-admin": {
		// Business management without low-level system config
		"User Management",
		"Employee Management",
		"Organization Read",
		"Group Management",
		"Group Read",
		"Content Management",
		"Content Read",
		"Comment Read",
		"Workflow Read",
		"Task Management",
		"Realtime Access",
		"Resource Read",
		"Payment Read",
		"CMS Management",
		"CMS Read",
		"Dashboard Access",
		"Profile Management",
		"Account Management",
		"Permission Read",
		"Dictionary Read",
		"System Read",
	},
	"manager": {
		// Department level management
		"User Read",
		"Employee Read",
		"Employee Update",
		"Employee Create",
		"Group Read",
		"Organization Read",
		"Content Read",
		"Comment Read",
		"Workflow Read",
		"Task Management",
		"Realtime Access",
		"Resource Read",
		"Dashboard Access",
		"Profile Management",
		"Account Management",
		"Dictionary Read",
		"System Read",
	},
	"employee": {
		// Basic employee access
		"User Read",
		"Employee Read",
		"Group Read",
		"Organization Read",
		"Content Read",
		"Comment Read",
		"Workflow Read",
		"Realtime Access",
		"Resource Read",
		"Dashboard Access",
		"Profile Management",
		"Account Management",
		"Dictionary Read",
		"System Read",
	},
	"guest": {
		// Minimal read-only access
		"Dashboard Access",
		"Content Read",
		"Profile Management",
		"Dictionary Read",
		"System Read",
	},
}

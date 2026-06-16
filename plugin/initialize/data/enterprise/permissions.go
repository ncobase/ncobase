package enterprise

import accessStructs "ncobase/core/access/structs"

// SystemDefaultPermissions defines simplified enterprise permissions
var SystemDefaultPermissions = []accessStructs.CreatePermissionBody{
	// Super admin permission
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Super Admin Access",
			Action:      "*",
			Subject:     "*",
			Description: "Super administrator wildcard permission",
		},
	},

	// Basic access permissions (all users need these)
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Dashboard Access",
			Action:      "read",
			Subject:     "dashboard",
			Description: "Access to dashboard and basic analytics",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Profile Management",
			Action:      "manage",
			Subject:     "profile",
			Description: "Manage own user profile",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Account Access",
			Action:      "read",
			Subject:     "account",
			Description: "Access to account information",
		},
	},

	// System management permissions
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
			Name:        "Menu Access",
			Action:      "read",
			Subject:     "menu",
			Description: "Access to system menus and navigation",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Dictionary Access",
			Action:      "read",
			Subject:     "dictionary",
			Description: "Access to system dictionaries",
		},
	},

	// User and Employee management
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "User Management",
			Action:      "manage",
			Subject:     "user",
			Description: "Full user management access",
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
			Description: "Full employee record management",
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

	// Organization management
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
			Name:        "Organization Read",
			Action:      "read",
			Subject:     "organization",
			Description: "View organizational information",
		},
	},

	// Content management
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Content Management",
			Action:      "manage",
			Subject:     "content",
			Description: "Full content management access",
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

	// Module access permissions
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Workflow Access",
			Action:      "read",
			Subject:     "workflow",
			Description: "Access to workflow and processes",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Resource Access",
			Action:      "read",
			Subject:     "resource",
			Description: "Access to resources and files",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Payment Read",
			Action:      "read",
			Subject:     "payments",
			Description: "View payment orders, provider metadata, and statistics",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Payment Management",
			Action:      "manage",
			Subject:     "payments",
			Description: "Manage payment products, channels, subscriptions, and order operations",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Payment Refund",
			Action:      "refund",
			Subject:     "payments",
			Description: "Process payment refunds",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Payment Administration",
			Action:      "admin",
			Subject:     "payments",
			Description: "Access payment logs and sensitive payment administration data",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Realtime Access",
			Action:      "read",
			Subject:     "realtime",
			Description: "Access realtime events, notifications, and personal channels",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Realtime Management",
			Action:      "manage",
			Subject:     "realtime",
			Description: "Manage realtime channels, events, and system notifications",
		},
	},
}

// RolePermissionMapping defines simplified role-permission relationships
var RolePermissionMapping = map[string][]string{
	"super-admin": {
		"Super Admin Access",
	},
	"system-admin": {
		"System Management",
		"User Management",
		"Employee Management",
		"Organization Management",
		"Content Management",
		"Dashboard Access",
		"Profile Management",
		"Account Access",
		"Menu Access",
		"Dictionary Access",
		"Workflow Access",
		"Resource Access",
		"Payment Read",
		"Payment Management",
		"Payment Refund",
		"Payment Administration",
		"Realtime Access",
		"Realtime Management",
	},
	"enterprise-admin": {
		"User Management",
		"Employee Management",
		"Organization Management",
		"Content Management",
		"Dashboard Access",
		"Profile Management",
		"Account Access",
		"Menu Access",
		"Dictionary Access",
		"Workflow Access",
		"Resource Access",
		"Payment Read",
		"Realtime Access",
		"Realtime Management",
	},
	"department-manager": {
		"User Read",
		"Employee Management",
		"Organization Read",
		"Content Read",
		"Dashboard Access",
		"Profile Management",
		"Account Access",
		"Menu Access",
		"Dictionary Access",
		"Workflow Access",
		"Resource Access",
		"Realtime Access",
	},
	"team-leader": {
		"User Read",
		"Employee Read",
		"Organization Read",
		"Content Read",
		"Dashboard Access",
		"Profile Management",
		"Account Access",
		"Menu Access",
		"Dictionary Access",
		"Workflow Access",
		"Resource Access",
		"Realtime Access",
	},
	"employee": {
		"User Read",
		"Employee Read",
		"Organization Read",
		"Content Read",
		"Dashboard Access",
		"Profile Management",
		"Account Access",
		"Menu Access",
		"Dictionary Access",
		"Workflow Access",
		"Resource Access",
		"Realtime Access",
	},
	"contractor": {
		"Dashboard Access",
		"Profile Management",
		"Account Access",
		"Menu Access",
		"Dictionary Access",
		"Content Read",
		"Resource Access",
	},
}

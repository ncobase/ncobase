package company

import accessStructs "ncobase/core/access/structs"

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
			Name:        "NCore Management",
			Action:      "manage",
			Subject:     "ncore",
			Description: "Manage NCore runtime extension surfaces",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Builder Management",
			Action:      "manage",
			Subject:     "builder",
			Description: "Access builder design and generation tools",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "IAM Management",
			Action:      "manage",
			Subject:     "iam",
			Description: "Manage identity and access administration surfaces",
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
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Dictionary Management",
			Action:      "manage",
			Subject:     "dictionary",
			Description: "Create, update, and delete dictionary data",
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

	// Space management
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Space Read",
			Action:      "read",
			Subject:     "space",
			Description: "View spaces and space membership data",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Space Management",
			Action:      "manage",
			Subject:     "space",
			Description: "Space management",
		},
	},

	// Organization management
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Organization Management",
			Action:      "manage",
			Subject:     "organization",
			Description: "Manage organization structure and members",
		},
	},
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
			Name:        "Plugin Management",
			Action:      "manage",
			Subject:     "plugins",
			Description: "Manage plugin administration and internal plugin surfaces",
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
			Name:        "Realtime Management",
			Action:      "manage",
			Subject:     "realtime",
			Description: "Manage realtime channels, events, and system notifications",
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
			Name:        "Resource Administration",
			Action:      "admin",
			Subject:     "resources",
			Description: "Access resource administration, quota, cleanup, and storage operations",
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
			Subject:     "payments",
			Description: "View payment information",
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
			Name:        "AI Read",
			Action:      "read",
			Subject:     "ai",
			Description: "View AI status, actions, usage, and own run history",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "AI Use",
			Action:      "use",
			Subject:     "ai",
			Description: "Use AI completion, embedding, streaming, and business action endpoints",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "AI Management",
			Action:      "manage",
			Subject:     "ai",
			Description: "Manage AI operational checks and cross-user run visibility",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "AI Administration",
			Action:      "admin",
			Subject:     "ai",
			Description: "Administer AI audit, provider, and policy-sensitive operations",
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
			Name:        "Workflow Management",
			Action:      "manage",
			Subject:     "workflow",
			Description: "Manage workflow templates, rules, delegations, and operations",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Taxonomy Management",
			Action:      "manage",
			Subject:     "taxonomies",
			Description: "Manage content taxonomy configuration",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Approval Management",
			Action:      "manage",
			Subject:     "approvals",
			Description: "Manage content approval surfaces",
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
		"NCore Management",
		"Builder Management",
		"IAM Management",
		"Menu Management",
		"Dictionary Read",
		"Dictionary Management",
		"User Management",
		"User Create",
		"Employee Management",
		"Role Management",
		"Permission Management",
		"Permission Read",
		"Space Read",
		"Space Management",
		"Organization Management",
		"Organization Read",
		"Group Management",
		"Group Read",
		"Content Management",
		"Content Read",
		"Comment Read",
		"TBP Management",
		"Plugin Management",
		"Realtime Access",
		"Realtime Management",
		"Resource Management",
		"Resource Administration",
		"Resource Read",
		"Payment Read",
		"Payment Management",
		"Payment Refund",
		"Payment Administration",
		"AI Read",
		"AI Use",
		"AI Management",
		"AI Administration",
		"Workflow Read",
		"Workflow Management",
		"Taxonomy Management",
		"Approval Management",
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
		"User Create",
		"Employee Management",
		"Space Read",
		"Organization Management",
		"Organization Read",
		"Group Management",
		"Group Read",
		"Content Management",
		"Content Read",
		"Comment Read",
		"Workflow Read",
		"Workflow Management",
		"Task Management",
		"Realtime Access",
		"Resource Read",
		"Payment Read",
		"AI Read",
		"AI Use",
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
		"AI Read",
		"AI Use",
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
		"AI Read",
		"AI Use",
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

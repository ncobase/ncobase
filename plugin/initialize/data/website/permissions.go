package website

import accessStructs "ncobase/core/access/structs"

// SystemDefaultPermissions for regular websites
var SystemDefaultPermissions = []accessStructs.CreatePermissionBody{
	// Super admin
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Super Admin Access",
			Action:      "*",
			Subject:     "*",
			Description: "Super admin wildcard permission",
		},
	},

	// Basic access
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

	// System management
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "System Management",
			Action:      "manage",
			Subject:     "system",
			Description: "System management",
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
			Description: "Menu management",
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

	// User management
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "User Management",
			Action:      "manage",
			Subject:     "user",
			Description: "User management",
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

	// Role management
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
			Name:        "Comment Management",
			Action:      "manage",
			Subject:     "comment",
			Description: "Manage comments",
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

	// Module permissions (simplified)
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
			Name:        "Resource Management",
			Action:      "manage",
			Subject:     "resource",
			Description: "Manage resources",
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
}

// RolePermissionMapping for websites
var RolePermissionMapping = map[string][]string{
	"super-admin": {
		"Super Admin Access",
	},
	"admin": {
		"System Management",
		"System Read",
		"Menu Management",
		"Dictionary Read",
		"User Management",
		"Role Management",
		"Permission Read",
		"Content Management",
		"Content Read",
		"Comment Management",
		"Comment Read",
		"Resource Management",
		"Resource Read",
		"Payment Read",
		"Payment Management",
		"Payment Refund",
		"Payment Administration",
		"Realtime Access",
		"Realtime Management",
		"AI Read",
		"AI Use",
		"AI Management",
		"AI Administration",
		"Dashboard Access",
		"Profile Management",
		"Account Management",
	},
	"manager": {
		"User Read",
		"Content Management",
		"Content Read",
		"Comment Management",
		"Comment Read",
		"Resource Read",
		"Payment Read",
		"Realtime Access",
		"AI Read",
		"AI Use",
		"Dashboard Access",
		"Profile Management",
		"Account Management",
		"Dictionary Read",
		"System Read",
	},
	"member": {
		"Content Read",
		"Comment Read",
		"Resource Read",
		"Realtime Access",
		"AI Read",
		"AI Use",
		"Dashboard Access",
		"Profile Management",
		"Account Management",
		"Dictionary Read",
		"System Read",
	},
	"viewer": {
		"Content Read",
		"Dashboard Access",
		"AI Read",
		"Profile Management",
		"Dictionary Read",
		"System Read",
	},
}

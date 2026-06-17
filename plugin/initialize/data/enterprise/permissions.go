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
			Description: "Manage system settings and configuration",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "System Administration",
			Action:      "admin",
			Subject:     "system",
			Description: "Access high-risk system administration, health, metrics, and diagnostics endpoints",
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
			Name:        "Menu Access",
			Action:      "read",
			Subject:     "menu",
			Description: "Access to system menus and navigation",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Menu Management",
			Action:      "manage",
			Subject:     "menu",
			Description: "Manage system menu structure",
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
			Description: "Update user profiles, status, credentials, and user-scoped settings",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "User Delete",
			Action:      "delete",
			Subject:     "user",
			Description: "Delete users and user-owned administrative API keys",
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
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Employee Create",
			Action:      "create",
			Subject:     "employee",
			Description: "Create employee records",
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
			Name:        "HR Management",
			Action:      "manage",
			Subject:     "hr",
			Description: "Manage HR administration surfaces and employee lifecycle operations",
		},
	},

	// Role, permission, and space administration
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
			Name:        "Role Read",
			Action:      "read",
			Subject:     "role",
			Description: "View roles, role permissions, and role user assignments",
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
			Description: "Manage spaces, memberships, and space configuration",
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
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Comment Read",
			Action:      "read",
			Subject:     "comment",
			Description: "View comments",
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
			Name:        "Workflow Management",
			Action:      "manage",
			Subject:     "workflow",
			Description: "Manage workflow templates, rules, delegations, and operations",
		},
	},
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
			Name:        "Resource Access",
			Action:      "read",
			Subject:     "resource",
			Description: "Access to resources and files",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Resource Management",
			Action:      "manage",
			Subject:     "resource",
			Description: "Manage resources and file lifecycle operations",
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
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Realtime Administration",
			Action:      "admin",
			Subject:     "realtime",
			Description: "Administer realtime channels, system notifications, and event operations",
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

// RolePermissionMapping defines simplified role-permission relationships
var RolePermissionMapping = map[string][]string{
	"super-admin": {
		"Super Admin Access",
	},
	"system-admin": {
		"System Management",
		"System Administration",
		"System Read",
		"NCore Management",
		"Builder Management",
		"IAM Management",
		"Menu Management",
		"User Management",
		"Employee Management",
		"Employee Create",
		"Employee Update",
		"HR Management",
		"Role Management",
		"Role Read",
		"Permission Management",
		"Permission Read",
		"Space Read",
		"Space Management",
		"Organization Management",
		"Organization Read",
		"Content Management",
		"Content Read",
		"Comment Read",
		"Taxonomy Management",
		"Approval Management",
		"Dashboard Access",
		"Profile Management",
		"Account Access",
		"Account Management",
		"Menu Access",
		"Dictionary Access",
		"Dictionary Management",
		"User Create",
		"User Update",
		"User Delete",
		"Workflow Access",
		"Workflow Management",
		"TBP Management",
		"Plugin Management",
		"Resource Access",
		"Resource Management",
		"Resource Administration",
		"CMS Read",
		"CMS Management",
		"Payment Read",
		"Payment Management",
		"Payment Refund",
		"Payment Administration",
		"Realtime Access",
		"Realtime Management",
		"Realtime Administration",
		"AI Read",
		"AI Use",
		"AI Management",
		"AI Administration",
	},
	"enterprise-admin": {
		"User Management",
		"User Create",
		"User Update",
		"User Delete",
		"Employee Management",
		"Space Read",
		"Organization Management",
		"Organization Read",
		"Content Management",
		"Content Read",
		"Comment Read",
		"Taxonomy Management",
		"Dashboard Access",
		"Profile Management",
		"Account Access",
		"Account Management",
		"Menu Access",
		"Dictionary Access",
		"Workflow Access",
		"Workflow Management",
		"Resource Access",
		"Resource Management",
		"CMS Read",
		"CMS Management",
		"Payment Read",
		"Realtime Access",
		"Realtime Management",
		"AI Read",
		"AI Use",
		"AI Management",
		"Role Read",
	},
	"department-manager": {
		"User Read",
		"Employee Management",
		"Organization Read",
		"Content Read",
		"Comment Read",
		"Dashboard Access",
		"Profile Management",
		"Account Access",
		"Account Management",
		"Menu Access",
		"Dictionary Access",
		"Workflow Access",
		"Resource Access",
		"Realtime Access",
		"AI Read",
		"AI Use",
	},
	"team-leader": {
		"User Read",
		"Employee Read",
		"Organization Read",
		"Content Read",
		"Comment Read",
		"Dashboard Access",
		"Profile Management",
		"Account Access",
		"Account Management",
		"Menu Access",
		"Dictionary Access",
		"Workflow Access",
		"Resource Access",
		"Realtime Access",
		"AI Read",
		"AI Use",
	},
	"employee": {
		"User Read",
		"Employee Read",
		"Organization Read",
		"Content Read",
		"Comment Read",
		"Dashboard Access",
		"Profile Management",
		"Account Access",
		"Account Management",
		"Menu Access",
		"Dictionary Access",
		"Workflow Access",
		"Resource Access",
		"Realtime Access",
		"AI Read",
		"AI Use",
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

package enterprise

import accessStructs "ncobase/access/structs"

// SystemDefaultPermissions defines enterprise system permissions
var SystemDefaultPermissions = []accessStructs.CreatePermissionBody{
	// ========== Super Administrator Permissions ==========
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Super Admin Access",
			Action:      "*",
			Subject:     "*",
			Description: "Super administrator wildcard permission",
		},
	},

	// ========== Basic Access Permissions (All Users) ==========
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
			Name:        "Account Access",
			Action:      "read",
			Subject:     "account",
			Description: "Access to user account information",
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
			Name:        "Dashboard Access",
			Action:      "read",
			Subject:     "dashboard",
			Description: "Access to dashboard and basic analytics",
		},
	},

	// ========== System Management Permissions ==========
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
			Name:        "Menu Management",
			Action:      "manage",
			Subject:     "menu",
			Description: "Manage system menus and navigation structure",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Dictionary Management",
			Action:      "manage",
			Subject:     "dictionary",
			Description: "Manage system dictionaries and enumerations",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Options Management",
			Action:      "manage",
			Subject:     "options",
			Description: "Manage system options and configurations",
		},
	},

	// ========== Tenant Management Permissions ==========
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
			Name:        "Tenant Read",
			Action:      "read",
			Subject:     "tenant",
			Description: "View tenant information",
		},
	},

	// ========== User Management Permissions ==========
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

	// ========== Employee Management Permissions ==========
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

	// ========== Organization Structure Permissions ==========
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
			Name:        "Team Management",
			Action:      "manage",
			Subject:     "team",
			Description: "Full team management access",
		},
	},

	// ========== Permission and Role Management ==========
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
			Name:        "Role Read",
			Action:      "read",
			Subject:     "role",
			Description: "View roles and role assignments",
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
			Name:        "Permission Read",
			Action:      "read",
			Subject:     "permission",
			Description: "View permissions",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Policy Management",
			Action:      "manage",
			Subject:     "policy",
			Description: "Manage Casbin policies",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Role Permission Management",
			Action:      "read",
			Subject:     "role_permission",
			Description: "View role permission assignments",
		},
	},

	// ========== Business Function Permissions ==========
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

	// ========== Content Management Permissions ==========
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
			Name:        "Topic Management",
			Action:      "manage",
			Subject:     "topic",
			Description: "Manage topics and content",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Taxonomy Management",
			Action:      "manage",
			Subject:     "taxonomy",
			Description: "Manage content taxonomies",
		},
	},

	// ========== Resource Management Permissions ==========
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Resource Management",
			Action:      "manage",
			Subject:     "resource",
			Description: "Full resource and file management",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Resource Read",
			Action:      "read",
			Subject:     "resource",
			Description: "View and download resources",
		},
	},

	// ========== Workflow Permissions ==========
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Workflow Management",
			Action:      "manage",
			Subject:     "workflow",
			Description: "Full workflow management access",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Workflow Read",
			Action:      "read",
			Subject:     "workflow",
			Description: "View workflows and processes",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Task Management",
			Action:      "manage",
			Subject:     "task",
			Description: "Manage workflow tasks",
		},
	},

	// ========== Payment Management Permissions ==========
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Payment Management",
			Action:      "manage",
			Subject:     "payment",
			Description: "Full payment system management",
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

	// ========== Real-time Communication Permissions ==========
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Realtime Access",
			Action:      "read",
			Subject:     "realtime",
			Description: "Access real-time features and notifications",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Notification Management",
			Action:      "manage",
			Subject:     "notification",
			Description: "Manage notifications and messaging",
		},
	},

	// ========== Proxy and Counter Permissions ==========
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Proxy Management",
			Action:      "manage",
			Subject:     "proxy",
			Description: "Manage proxy configurations",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Counter Management",
			Action:      "manage",
			Subject:     "counter",
			Description: "Manage counters and metrics",
		},
	},
}

// RolePermissionMapping defines role-permission relationships
var RolePermissionMapping = map[string][]string{
	// System level roles
	"super-admin": {
		"Super Admin Access", // Wildcard permission
	},
	"system-admin": {
		// System management
		"System Management",
		"Menu Management",
		"Dictionary Management",
		"Options Management",
		"Tenant Management",
		"User Management",
		"Employee Management",
		"Organization Management",
		"Group Management",
		"Department Management",
		"Team Management",
		"Role Management",
		"Permission Management",
		"Policy Management",
		"Role Permission Management",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		"Enterprise Overview",
		"Cross Company Access",
		// Business functions
		"Content Management",
		"Topic Management",
		"Taxonomy Management",
		"Resource Management",
		"Workflow Management",
		"Task Management",
		"Payment Management",
		"Notification Management",
		"Realtime Access",
		"Proxy Management",
		"Counter Management",
	},

	// Enterprise level roles
	"enterprise-admin": {
		// Enterprise management
		"User Management",
		"Employee Management",
		"Organization Management",
		"Group Management",
		"Department Management",
		"Team Management",
		"Financial Management",
		"HR Management",
		"Role Read",
		"Permission Read",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		"Enterprise Overview",
		"Cross Company Access",
		"Tenant Read",
		// Business functions
		"Content Management",
		"Topic Management",
		"Taxonomy Management",
		"Resource Management",
		"Workflow Management",
		"Task Management",
		"Realtime Access",
		"Notification Management",
	},
	"tenant-admin": {
		// Tenant management
		"User Read",
		"User Create",
		"User Update",
		"Employee Management",
		"Group Read",
		"Department Management",
		"Team Management",
		"Role Read",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		"Tenant Read",
		// Business functions
		"Content Read",
		"Resource Read",
		"Workflow Read",
		"Task Management",
		"Realtime Access",
	},

	// Functional department roles
	"hr-manager": {
		// HR management
		"Employee Management",
		"HR Management",
		"User Read",
		"User Create",
		"User Update",
		"Group Read",
		"Department Management",
		"Team Management",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Related business
		"Workflow Read",
		"Task Management",
		"Realtime Access",
	},
	"finance-manager": {
		// Finance management
		"Financial Management",
		"Employee Read",
		"User Read",
		"Group Read",
		"Payment Management",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Related business
		"Workflow Read",
		"Task Management",
		"Realtime Access",
	},
	"it-manager": {
		// IT management
		"System Management",
		"User Management",
		"Employee Management",
		"Organization Management",
		"Group Management",
		"Department Management",
		"Team Management",
		"Role Read",
		"Permission Read",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		"Enterprise Overview",
		// IT related business
		"Content Management",
		"Resource Management",
		"Workflow Management",
		"Realtime Access",
		"Notification Management",
		"Proxy Management",
		"Counter Management",
	},

	// Management level roles
	"department-manager": {
		// Department management
		"Employee Read",
		"Employee Update",
		"User Read",
		"Department Management",
		"Team Management",
		"Group Read",
		"Task Management",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Related business
		"Content Read",
		"Resource Read",
		"Workflow Read",
		"Realtime Access",
	},
	"team-leader": {
		// Team leadership
		"Employee Read",
		"User Read",
		"Group Read",
		"Task Management",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Related business
		"Content Read",
		"Resource Read",
		"Workflow Read",
		"Realtime Access",
	},
	"project-manager": {
		// Project management
		"Employee Read",
		"Employee Update",
		"User Read",
		"Team Management",
		"Group Read",
		"Task Management",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Project related
		"Content Management",
		"Resource Management",
		"Workflow Management",
		"Realtime Access",
	},

	// Professional technical roles
	"technical-lead": {
		// Technical management
		"Employee Read",
		"Employee Update",
		"User Read",
		"Department Management",
		"Team Management",
		"Group Read",
		"Task Management",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Technical business
		"Content Management",
		"Resource Management",
		"Workflow Management",
		"Realtime Access",
		"Proxy Management",
		"Counter Management",
	},
	"developer": {
		// Developer permissions
		"User Read",
		"Employee Read",
		"Group Read",
		"Task Management",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Development related
		"Content Read",
		"Resource Management",
		"Workflow Read",
		"Realtime Access",
	},
	"designer": {
		// Designer permissions
		"User Read",
		"Employee Read",
		"Group Read",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Design related
		"Content Management",
		"Resource Management",
		"Workflow Read",
		"Realtime Access",
	},
	"content-creator": {
		// Content creation
		"User Read",
		"Employee Read",
		"Group Read",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Content related
		"Content Management",
		"Topic Management",
		"Taxonomy Management",
		"Resource Management",
		"Workflow Read",
		"Realtime Access",
	},
	"marketing-specialist": {
		// Marketing specialist
		"User Read",
		"Employee Read",
		"Group Read",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Marketing related
		"Content Management",
		"Resource Management",
		"Workflow Read",
		"Realtime Access",
	},

	// Quality assurance roles
	"qa-manager": {
		// QA management
		"Employee Read",
		"Employee Update",
		"User Read",
		"Team Management",
		"Group Read",
		"Task Management",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// QA related
		"Content Read",
		"Resource Read",
		"Workflow Management",
		"Realtime Access",
	},
	"qa-tester": {
		// QA testing
		"User Read",
		"Employee Read",
		"Group Read",
		"Task Management",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Testing related
		"Content Read",
		"Resource Read",
		"Workflow Read",
		"Realtime Access",
	},

	// Data and analysis roles
	"data-analyst": {
		// Data analysis
		"User Read",
		"Employee Read",
		"Group Read",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		"Enterprise Overview",
		// Analysis related
		"Content Read",
		"Resource Read",
		"Workflow Read",
		"Payment Read",
		"Realtime Access",
	},
	"business-analyst": {
		// Business analysis
		"User Read",
		"Employee Read",
		"Group Read",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		"Enterprise Overview",
		// Business analysis
		"Content Read",
		"Resource Read",
		"Workflow Management",
		"Realtime Access",
	},

	// Customer service roles
	"customer-service-manager": {
		// Customer service management
		"Employee Read",
		"Employee Update",
		"User Read",
		"Team Management",
		"Group Read",
		"Task Management",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Customer service related
		"Content Read",
		"Resource Read",
		"Workflow Management",
		"Notification Management",
		"Realtime Access",
	},
	"customer-service-rep": {
		// Customer service representative
		"User Read",
		"Employee Read",
		"Group Read",
		"Task Management",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Customer service related
		"Content Read",
		"Resource Read",
		"Workflow Read",
		"Realtime Access",
	},

	// Basic employee roles
	"senior-employee": {
		// Senior employee
		"User Read",
		"Employee Read",
		"Group Read",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Extended business access
		"Content Read",
		"Resource Management",
		"Workflow Read",
		"Task Management",
		"Realtime Access",
	},
	"employee": {
		// Standard employee
		"User Read",
		"Employee Read",
		"Group Read",
		// Basic access (mandatory for all users)
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Basic business
		"Content Read",
		"Resource Read",
		"Workflow Read",
		"Realtime Access",
	},
	"junior-employee": {
		// Junior employee
		"User Read",
		"Employee Read",
		"Group Read",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Basic business (limited)
		"Content Read",
		"Resource Read",
		"Workflow Read",
		"Realtime Access",
	},

	// Special roles
	"contractor": {
		// External contractor (limited)
		"User Read",
		"Group Read",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Basic business (limited)
		"Content Read",
		"Resource Read",
		"Realtime Access",
	},
	"consultant": {
		// External consultant
		"User Read",
		"Employee Read",
		"Group Read",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Consultant related
		"Content Read",
		"Resource Read",
		"Workflow Read",
		"Realtime Access",
	},
	"intern": {
		// Intern (minimal permissions)
		"User Read",
		"Group Read",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Basic business (limited)
		"Content Read",
		"Resource Read",
		"Realtime Access",
	},
	"auditor": {
		// External auditor (read-only)
		"User Read",
		"Employee Read",
		"Group Read",
		"Role Read",
		"Permission Read",
		"Tenant Read",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		"Enterprise Overview",
		// Audit related (read-only)
		"Content Read",
		"Resource Read",
		"Workflow Read",
		"Payment Read",
	},

	// Organizational structure roles
	"enterprise-executive": {
		// Enterprise executive
		"User Management",
		"Employee Management",
		"Organization Management",
		"Group Management",
		"Department Management",
		"Team Management",
		"Financial Management",
		"HR Management",
		"Role Read",
		"Permission Read",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		"Enterprise Overview",
		"Cross Company Access",
		"Tenant Read",
		// Business functions
		"Content Management",
		"Resource Management",
		"Workflow Management",
		"Payment Read",
		"Realtime Access",
		"Notification Management",
	},
	"company-director": {
		// Company director
		"Employee Management",
		"Organization Management",
		"Department Management",
		"Team Management",
		"User Read",
		"User Create",
		"User Update",
		"Group Management",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		"Enterprise Overview",
		// Business functions
		"Content Management",
		"Resource Management",
		"Workflow Management",
		"Task Management",
		"Realtime Access",
	},
	"department-head": {
		// Department head
		"Employee Read",
		"Employee Update",
		"Department Management",
		"Team Management",
		"User Read",
		"Group Read",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Business functions
		"Content Read",
		"Resource Management",
		"Workflow Read",
		"Task Management",
		"Realtime Access",
	},
	"team-supervisor": {
		// Team supervisor
		"Employee Read",
		"User Read",
		"Group Read",
		"Task Management",
		// Basic access
		"Menu Access",
		"Account Access",
		"Profile Management",
		"Dashboard Access",
		// Business functions
		"Content Read",
		"Resource Read",
		"Workflow Read",
		"Realtime Access",
	},
}

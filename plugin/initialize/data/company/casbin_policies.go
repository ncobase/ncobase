package company

// CasbinPolicyRules aligned with actual routes and menu structure
var CasbinPolicyRules = [][]string{
	// Super Admin - full access
	{"super-admin", "*", "*", "*"},

	// System Admin - comprehensive system management
	{"system-admin", "*", "/", "GET"},
	{"system-admin", "*", "/swagger/*", "GET"},

	// System module routes
	{"system-admin", "*", "/sys/*", "*"},

	// User management routes
	{"system-admin", "*", "/user/*", "*"}, // Note: should be /sys/users based on menu

	// TBP module routes
	{"system-admin", "*", "/tbp/*", "*"},

	// Realtime module routes
	{"system-admin", "*", "/rt/*", "*"},

	// Resources module routes
	{"system-admin", "*", "/res/*", "*"},

	// Payment module routes
	{"system-admin", "*", "/pay/*", "*"},

	// Workflow module routes
	{"system-admin", "*", "/flow/*", "*"},

	// IAM module routes
	{"system-admin", "*", "/iam/*", "*"},

	// Organization module routes
	{"system-admin", "*", "/org/*", "*"},

	// CMS module routes
	{"system-admin", "*", "/cms/*", "*"},

	// Plugins module routes
	{"system-admin", "*", "/plug/*", "*"},

	// Enterprise Admin - business management
	{"enterprise-admin", "*", "/", "GET"},
	{"enterprise-admin", "*", "/iam/account", "GET"},
	{"enterprise-admin", "*", "/iam/account/tenant", "GET"},
	{"enterprise-admin", "*", "/iam/account/tenants", "GET"},

	// User and employee management
	{"enterprise-admin", "*", "/sys/users", "*"}, // Aligned with menu path
	{"enterprise-admin", "*", "/sys/employees", "*"},

	// Limited system access (read-only)
	{"enterprise-admin", "*", "/sys/menus", "GET"},
	{"enterprise-admin", "*", "/sys/dictionaries", "GET"},
	{"enterprise-admin", "*", "/sys/options", "GET"},

	// Organization management
	{"enterprise-admin", "*", "/org/groups", "*"},
	{"enterprise-admin", "*", "/sys/group", "*"}, // System group management

	// Content management
	{"enterprise-admin", "*", "/cms/topics", "*"},
	{"enterprise-admin", "*", "/cms/taxonomies", "*"},
	{"enterprise-admin", "*", "/cms/channels", "*"},
	{"enterprise-admin", "*", "/cms/media", "*"},

	// Workflow management
	{"enterprise-admin", "*", "/flow/processes", "*"},
	{"enterprise-admin", "*", "/flow/tasks", "*"},

	// Limited realtime access
	{"enterprise-admin", "*", "/rt/notifications", "*"},
	{"enterprise-admin", "*", "/rt/channels", "GET"},

	// Resource access
	{"enterprise-admin", "*", "/res", "GET"},
	{"enterprise-admin", "*", "/res/search", "GET"},

	// Payment viewing
	{"enterprise-admin", "*", "/pay/orders", "GET"},

	// Manager - departmental management
	{"manager", "*", "/", "GET"},
	{"manager", "*", "/iam/account", "GET"},
	{"manager", "*", "/iam/account/tenant", "GET"},
	{"manager", "*", "/iam/account/tenants", "GET"},

	// User and employee access (read + limited update)
	{"manager", "*", "/sys/users", "GET"},
	{"manager", "*", "/sys/employees", "GET"},
	{"manager", "*", "/sys/employees", "POST"},  // Can create employees
	{"manager", "*", "/sys/employees/*", "PUT"}, // Can update employees

	// Basic system information
	{"manager", "*", "/sys/menus", "GET"},
	{"manager", "*", "/sys/dictionaries", "GET"},
	{"manager", "*", "/sys/options", "GET"},

	// Organization viewing
	{"manager", "*", "/org/groups", "GET"},
	{"manager", "*", "/sys/group", "GET"},

	// Content access
	{"manager", "*", "/cms/topics", "GET"},

	// Task management
	{"manager", "*", "/flow/tasks", "*"},

	// Notifications
	{"manager", "*", "/rt/notifications", "GET"},

	// Resource access
	{"manager", "*", "/res", "GET"},

	// Employee - basic access
	{"employee", "*", "/", "GET"},
	{"employee", "*", "/iam/account", "GET"},
	{"employee", "*", "/iam/account/tenant", "GET"},
	{"employee", "*", "/iam/account/tenants", "GET"},

	// Basic system information (read-only)
	{"employee", "*", "/sys/menus", "GET"},
	{"employee", "*", "/sys/dictionaries", "GET"},
	{"employee", "*", "/sys/options", "GET"},

	// User and employee information (read-only)
	{"employee", "*", "/sys/users", "GET"},
	{"employee", "*", "/sys/employees", "GET"},

	// Organization information (read-only)
	{"employee", "*", "/org/groups", "GET"},
	{"employee", "*", "/sys/group", "GET"},

	// Content access (read-only)
	{"employee", "*", "/cms/topics", "GET"},

	// Task viewing
	{"employee", "*", "/flow/tasks", "GET"},

	// Notifications
	{"employee", "*", "/rt/notifications", "GET"},

	// Resource access
	{"employee", "*", "/res", "GET"},

	// Guest - minimal read-only access
	{"guest", "*", "/", "GET"},
	{"guest", "*", "/iam/account", "GET"},
	{"guest", "*", "/sys/menus", "GET"},
	{"guest", "*", "/sys/dictionaries", "GET"},
	{"guest", "*", "/cms/topics", "GET"},
	{"guest", "*", "/res", "GET"},
}

// RoleInheritanceRules defines role inheritance
var RoleInheritanceRules = [][]string{
	{"system-admin", "super-admin", "*"},
	{"enterprise-admin", "system-admin", "*"},
	{"manager", "enterprise-admin", "*"},
	{"employee", "manager", "*"},
	{"guest", "employee", "*"},
}

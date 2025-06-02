package company

// CasbinPolicyRules for 6-parameter model: sub, dom, obj, act, v4, v5
var CasbinPolicyRules = [][]string{
	// Super Admin - wildcard access (handled by matcher, but explicit rules for clarity)
	{"super-admin", "*", "*", "*", "", ""},

	// System Admin - comprehensive system management
	{"system-admin", "*", "/", "GET", "", ""},
	{"system-admin", "*", "/", "read", "", ""},
	{"system-admin", "*", "/swagger/*", "GET", "", ""},
	{"system-admin", "*", "/swagger/*", "read", "", ""},
	{"system-admin", "*", "/tbp/*", "*", "", ""},
	{"system-admin", "*", "/rt/*", "*", "", ""},
	{"system-admin", "*", "/res/*", "*", "", ""},
	{"system-admin", "*", "/pay/*", "*", "", ""},
	{"system-admin", "*", "/sys/*", "*", "", ""},
	{"system-admin", "*", "/plug/*", "*", "", ""},
	{"system-admin", "*", "/flow/*", "*", "", ""},
	{"system-admin", "*", "/iam/*", "*", "", ""},
	{"system-admin", "*", "/org/*", "*", "", ""},
	{"system-admin", "*", "/cms/*", "*", "", ""},

	// Company Admin - business management (both HTTP methods and semantic actions)
	{"company-admin", "*", "/", "GET", "", ""},
	{"company-admin", "*", "/", "read", "", ""},
	{"company-admin", "*", "/account", "GET", "", ""},
	{"company-admin", "*", "/account", "read", "", ""},
	{"company-admin", "*", "/account/tenant", "GET", "", ""},
	{"company-admin", "*", "/account/tenant", "read", "", ""},
	{"company-admin", "*", "/account/tenants", "GET", "", ""},
	{"company-admin", "*", "/account/tenants", "read", "", ""},

	// User management - semantic actions preferred
	{"company-admin", "*", "/sys/users", "read", "", ""},
	{"company-admin", "*", "/sys/users", "create", "", ""},
	{"company-admin", "*", "/sys/users", "update", "", ""},
	{"company-admin", "*", "/sys/users", "delete", "", ""},
	{"company-admin", "*", "/sys/users/*", "read", "", ""},
	{"company-admin", "*", "/sys/users/*", "update", "", ""},
	{"company-admin", "*", "/sys/users/*", "delete", "", ""},

	// Employee management
	{"company-admin", "*", "/sys/employees", "read", "", ""},
	{"company-admin", "*", "/sys/employees", "create", "", ""},
	{"company-admin", "*", "/sys/employees", "update", "", ""},
	{"company-admin", "*", "/sys/employees", "delete", "", ""},
	{"company-admin", "*", "/sys/employees/*", "read", "", ""},
	{"company-admin", "*", "/sys/employees/*", "update", "", ""},
	{"company-admin", "*", "/sys/employees/*", "delete", "", ""},

	// System info access
	{"company-admin", "*", "/sys/menus", "GET", "", ""},
	{"company-admin", "*", "/sys/menus", "read", "", ""},
	{"company-admin", "*", "/sys/dictionaries", "GET", "", ""},
	{"company-admin", "*", "/sys/dictionaries", "read", "", ""},
	{"company-admin", "*", "/sys/options", "GET", "", ""},
	{"company-admin", "*", "/sys/options", "read", "", ""},

	// Organization management
	{"company-admin", "*", "/org/groups", "read", "", ""},
	{"company-admin", "*", "/org/groups", "create", "", ""},
	{"company-admin", "*", "/org/groups", "update", "", ""},
	{"company-admin", "*", "/org/groups", "delete", "", ""},
	{"company-admin", "*", "/org/groups/*", "read", "", ""},
	{"company-admin", "*", "/org/groups/*", "update", "", ""},
	{"company-admin", "*", "/org/groups/*", "delete", "", ""},

	// Content management
	{"company-admin", "*", "/cms/topics", "read", "", ""},
	{"company-admin", "*", "/cms/topics", "create", "", ""},
	{"company-admin", "*", "/cms/topics", "update", "", ""},
	{"company-admin", "*", "/cms/topics", "delete", "", ""},
	{"company-admin", "*", "/cms/*", "read", "", ""},
	{"company-admin", "*", "/cms/*", "create", "", ""},
	{"company-admin", "*", "/cms/*", "update", "", ""},

	// Workflow and notifications
	{"company-admin", "*", "/flow/processes", "read", "", ""},
	{"company-admin", "*", "/flow/processes", "create", "", ""},
	{"company-admin", "*", "/flow/processes", "update", "", ""},
	{"company-admin", "*", "/flow/tasks", "read", "", ""},
	{"company-admin", "*", "/flow/tasks", "create", "", ""},
	{"company-admin", "*", "/flow/tasks", "update", "", ""},
	{"company-admin", "*", "/rt/notifications", "read", "", ""},
	{"company-admin", "*", "/rt/notifications", "create", "", ""},
	{"company-admin", "*", "/rt/notifications", "update", "", ""},
	{"company-admin", "*", "/rt/channels", "GET", "", ""},
	{"company-admin", "*", "/rt/channels", "read", "", ""},

	// Resource access
	{"company-admin", "*", "/res", "GET", "", ""},
	{"company-admin", "*", "/res", "read", "", ""},
	{"company-admin", "*", "/res/search", "GET", "", ""},
	{"company-admin", "*", "/res/search", "read", "", ""},

	// Payment viewing
	{"company-admin", "*", "/pay/orders", "GET", "", ""},
	{"company-admin", "*", "/pay/orders", "read", "", ""},

	// Manager - departmental management
	{"manager", "*", "/", "GET", "", ""},
	{"manager", "*", "/", "read", "", ""},
	{"manager", "*", "/account", "GET", "", ""},
	{"manager", "*", "/account", "read", "", ""},
	{"manager", "*", "/account/tenant", "GET", "", ""},
	{"manager", "*", "/account/tenant", "read", "", ""},
	{"manager", "*", "/account/tenants", "GET", "", ""},
	{"manager", "*", "/account/tenants", "read", "", ""},

	// User and employee access (read + limited update)
	{"manager", "*", "/sys/users", "GET", "", ""},
	{"manager", "*", "/sys/users", "read", "", ""},
	{"manager", "*", "/sys/employees", "GET", "", ""},
	{"manager", "*", "/sys/employees", "read", "", ""},
	{"manager", "*", "/sys/employees", "POST", "", ""},
	{"manager", "*", "/sys/employees", "create", "", ""},
	{"manager", "*", "/sys/employees/*", "PUT", "", ""},
	{"manager", "*", "/sys/employees/*", "update", "", ""},

	// System info
	{"manager", "*", "/sys/menus", "GET", "", ""},
	{"manager", "*", "/sys/menus", "read", "", ""},
	{"manager", "*", "/sys/dictionaries", "GET", "", ""},
	{"manager", "*", "/sys/dictionaries", "read", "", ""},
	{"manager", "*", "/sys/options", "GET", "", ""},
	{"manager", "*", "/sys/options", "read", "", ""},

	// Organization and content
	{"manager", "*", "/org/groups", "GET", "", ""},
	{"manager", "*", "/org/groups", "read", "", ""},
	{"manager", "*", "/cms/topics", "GET", "", ""},
	{"manager", "*", "/cms/topics", "read", "", ""},

	// Task management and notifications
	{"manager", "*", "/flow/tasks", "read", "", ""},
	{"manager", "*", "/flow/tasks", "create", "", ""},
	{"manager", "*", "/flow/tasks", "update", "", ""},
	{"manager", "*", "/rt/notifications", "GET", "", ""},
	{"manager", "*", "/rt/notifications", "read", "", ""},

	// Resource access
	{"manager", "*", "/res", "GET", "", ""},
	{"manager", "*", "/res", "read", "", ""},

	// Employee - basic access
	{"employee", "*", "/", "GET", "", ""},
	{"employee", "*", "/", "read", "", ""},
	{"employee", "*", "/account", "GET", "", ""},
	{"employee", "*", "/account", "read", "", ""},
	{"employee", "*", "/account/tenant", "GET", "", ""},
	{"employee", "*", "/account/tenant", "read", "", ""},
	{"employee", "*", "/account/tenants", "GET", "", ""},
	{"employee", "*", "/account/tenants", "read", "", ""},

	// Basic system info (read-only)
	{"employee", "*", "/sys/menus", "GET", "", ""},
	{"employee", "*", "/sys/menus", "read", "", ""},
	{"employee", "*", "/sys/dictionaries", "GET", "", ""},
	{"employee", "*", "/sys/dictionaries", "read", "", ""},
	{"employee", "*", "/sys/options", "GET", "", ""},
	{"employee", "*", "/sys/options", "read", "", ""},
	{"employee", "*", "/sys/users", "GET", "", ""},
	{"employee", "*", "/sys/users", "read", "", ""},
	{"employee", "*", "/sys/employees", "GET", "", ""},
	{"employee", "*", "/sys/employees", "read", "", ""},

	// Organization and content (read-only)
	{"employee", "*", "/org/groups", "GET", "", ""},
	{"employee", "*", "/org/groups", "read", "", ""},
	{"employee", "*", "/cms/topics", "GET", "", ""},
	{"employee", "*", "/cms/topics", "read", "", ""},

	// Task viewing and notifications
	{"employee", "*", "/flow/tasks", "GET", "", ""},
	{"employee", "*", "/flow/tasks", "read", "", ""},
	{"employee", "*", "/rt/notifications", "GET", "", ""},
	{"employee", "*", "/rt/notifications", "read", "", ""},

	// Resource access
	{"employee", "*", "/res", "GET", "", ""},
	{"employee", "*", "/res", "read", "", ""},

	// Guest - minimal read-only access
	{"guest", "*", "/", "GET", "", ""},
	{"guest", "*", "/", "read", "", ""},
	{"guest", "*", "/account", "GET", "", ""},
	{"guest", "*", "/account", "read", "", ""},
	{"guest", "*", "/sys/menus", "GET", "", ""},
	{"guest", "*", "/sys/menus", "read", "", ""},
	{"guest", "*", "/sys/dictionaries", "GET", "", ""},
	{"guest", "*", "/sys/dictionaries", "read", "", ""},
	{"guest", "*", "/cms/topics", "GET", "", ""},
	{"guest", "*", "/cms/topics", "read", "", ""},
	{"guest", "*", "/res", "GET", "", ""},
	{"guest", "*", "/res", "read", "", ""},
}

// RoleInheritanceRules for g matcher: child_role, parent_role, domain
var RoleInheritanceRules = [][]string{
	{"system-admin", "super-admin", "*"},
	{"company-admin", "system-admin", "*"},
	{"manager", "company-admin", "*"},
	{"employee", "manager", "*"},
	{"guest", "employee", "*"},
}

package enterprise

// CasbinPolicyRules for 6-parameter model: sub, dom, obj, act, v4, v5
var CasbinPolicyRules = [][]string{
	// Super Admin - wildcard access
	{"super-admin", "*", "*", "*", "", ""},

	// System Admin - full system management
	{"system-admin", "*", "/", "GET", "", ""},
	{"system-admin", "*", "/", "read", "", ""},
	{"system-admin", "*", "/swagger/*", "GET", "", ""},
	{"system-admin", "*", "/swagger/*", "read", "", ""},
	{"system-admin", "*", "/sys/*", "*", "", ""},
	{"system-admin", "*", "/tbp/*", "*", "", ""},
	{"system-admin", "*", "/rt/*", "*", "", ""},
	{"system-admin", "*", "/res/*", "*", "", ""},
	{"system-admin", "*", "/pay/*", "*", "", ""},
	{"system-admin", "*", "/flow/*", "*", "", ""},
	{"system-admin", "*", "/iam/*", "*", "", ""},
	{"system-admin", "*", "/org/*", "*", "", ""},
	{"system-admin", "*", "/cms/*", "*", "", ""},
	{"system-admin", "*", "/plug/*", "*", "", ""},

	// Enterprise Admin - business management
	{"enterprise-admin", "*", "/", "GET", "", ""},
	{"enterprise-admin", "*", "/", "read", "", ""},
	{"enterprise-admin", "*", "/account", "GET", "", ""},
	{"enterprise-admin", "*", "/account", "read", "", ""},
	{"enterprise-admin", "*", "/account/tenant", "GET", "", ""},
	{"enterprise-admin", "*", "/account/tenant", "read", "", ""},
	{"enterprise-admin", "*", "/account/tenants", "GET", "", ""},
	{"enterprise-admin", "*", "/account/tenants", "read", "", ""},

	// User and employee management
	{"enterprise-admin", "*", "/sys/users", "read", "", ""},
	{"enterprise-admin", "*", "/sys/users", "create", "", ""},
	{"enterprise-admin", "*", "/sys/users", "update", "", ""},
	{"enterprise-admin", "*", "/sys/users", "delete", "", ""},
	{"enterprise-admin", "*", "/sys/users/*", "read", "", ""},
	{"enterprise-admin", "*", "/sys/users/*", "update", "", ""},
	{"enterprise-admin", "*", "/sys/users/*", "delete", "", ""},
	{"enterprise-admin", "*", "/sys/employees", "read", "", ""},
	{"enterprise-admin", "*", "/sys/employees", "create", "", ""},
	{"enterprise-admin", "*", "/sys/employees", "update", "", ""},
	{"enterprise-admin", "*", "/sys/employees", "delete", "", ""},
	{"enterprise-admin", "*", "/sys/employees/*", "read", "", ""},
	{"enterprise-admin", "*", "/sys/employees/*", "update", "", ""},
	{"enterprise-admin", "*", "/sys/employees/*", "delete", "", ""},

	// System info access
	{"enterprise-admin", "*", "/sys/menus", "GET", "", ""},
	{"enterprise-admin", "*", "/sys/menus", "read", "", ""},
	{"enterprise-admin", "*", "/sys/dictionaries", "GET", "", ""},
	{"enterprise-admin", "*", "/sys/dictionaries", "read", "", ""},
	{"enterprise-admin", "*", "/sys/options", "GET", "", ""},
	{"enterprise-admin", "*", "/sys/options", "read", "", ""},

	// Organization management
	{"enterprise-admin", "*", "/org/groups", "read", "", ""},
	{"enterprise-admin", "*", "/org/groups", "create", "", ""},
	{"enterprise-admin", "*", "/org/groups", "update", "", ""},
	{"enterprise-admin", "*", "/org/groups", "delete", "", ""},
	{"enterprise-admin", "*", "/org/groups/*", "read", "", ""},
	{"enterprise-admin", "*", "/org/groups/*", "update", "", ""},
	{"enterprise-admin", "*", "/org/groups/*", "delete", "", ""},

	// Content and workflow management
	{"enterprise-admin", "*", "/cms/*", "read", "", ""},
	{"enterprise-admin", "*", "/cms/*", "create", "", ""},
	{"enterprise-admin", "*", "/cms/*", "update", "", ""},
	{"enterprise-admin", "*", "/cms/*", "delete", "", ""},
	{"enterprise-admin", "*", "/flow/processes", "read", "", ""},
	{"enterprise-admin", "*", "/flow/processes", "create", "", ""},
	{"enterprise-admin", "*", "/flow/processes", "update", "", ""},
	{"enterprise-admin", "*", "/flow/tasks", "read", "", ""},
	{"enterprise-admin", "*", "/flow/tasks", "create", "", ""},
	{"enterprise-admin", "*", "/flow/tasks", "update", "", ""},

	// Notifications and resources
	{"enterprise-admin", "*", "/rt/notifications", "read", "", ""},
	{"enterprise-admin", "*", "/rt/notifications", "create", "", ""},
	{"enterprise-admin", "*", "/rt/notifications", "update", "", ""},
	{"enterprise-admin", "*", "/rt/channels", "GET", "", ""},
	{"enterprise-admin", "*", "/rt/channels", "read", "", ""},
	{"enterprise-admin", "*", "/res", "GET", "", ""},
	{"enterprise-admin", "*", "/res", "read", "", ""},
	{"enterprise-admin", "*", "/res/search", "GET", "", ""},
	{"enterprise-admin", "*", "/res/search", "read", "", ""},
	{"enterprise-admin", "*", "/pay/orders", "GET", "", ""},
	{"enterprise-admin", "*", "/pay/orders", "read", "", ""},

	// Department Manager - department level management
	{"department-manager", "*", "/", "GET", "", ""},
	{"department-manager", "*", "/", "read", "", ""},
	{"department-manager", "*", "/account", "GET", "", ""},
	{"department-manager", "*", "/account", "read", "", ""},
	{"department-manager", "*", "/account/tenant", "GET", "", ""},
	{"department-manager", "*", "/account/tenant", "read", "", ""},

	// Employee management within department
	{"department-manager", "*", "/sys/users", "GET", "", ""},
	{"department-manager", "*", "/sys/users", "read", "", ""},
	{"department-manager", "*", "/sys/employees", "GET", "", ""},
	{"department-manager", "*", "/sys/employees", "read", "", ""},
	{"department-manager", "*", "/sys/employees", "POST", "", ""},
	{"department-manager", "*", "/sys/employees", "create", "", ""},
	{"department-manager", "*", "/sys/employees/*", "PUT", "", ""},
	{"department-manager", "*", "/sys/employees/*", "update", "", ""},

	// System info and organization
	{"department-manager", "*", "/sys/menus", "GET", "", ""},
	{"department-manager", "*", "/sys/menus", "read", "", ""},
	{"department-manager", "*", "/sys/dictionaries", "GET", "", ""},
	{"department-manager", "*", "/sys/dictionaries", "read", "", ""},
	{"department-manager", "*", "/org/groups", "GET", "", ""},
	{"department-manager", "*", "/org/groups", "read", "", ""},
	{"department-manager", "*", "/cms/topics", "GET", "", ""},
	{"department-manager", "*", "/cms/topics", "read", "", ""},

	// Task and notification management
	{"department-manager", "*", "/flow/tasks", "read", "", ""},
	{"department-manager", "*", "/flow/tasks", "create", "", ""},
	{"department-manager", "*", "/flow/tasks", "update", "", ""},
	{"department-manager", "*", "/rt/notifications", "GET", "", ""},
	{"department-manager", "*", "/rt/notifications", "read", "", ""},
	{"department-manager", "*", "/res", "GET", "", ""},
	{"department-manager", "*", "/res", "read", "", ""},

	// Team Leader - team level management
	{"team-leader", "*", "/", "GET", "", ""},
	{"team-leader", "*", "/", "read", "", ""},
	{"team-leader", "*", "/account", "GET", "", ""},
	{"team-leader", "*", "/account", "read", "", ""},
	{"team-leader", "*", "/account/tenant", "GET", "", ""},
	{"team-leader", "*", "/account/tenant", "read", "", ""},

	// Basic user and employee info
	{"team-leader", "*", "/sys/users", "GET", "", ""},
	{"team-leader", "*", "/sys/users", "read", "", ""},
	{"team-leader", "*", "/sys/employees", "GET", "", ""},
	{"team-leader", "*", "/sys/employees", "read", "", ""},
	{"team-leader", "*", "/sys/menus", "GET", "", ""},
	{"team-leader", "*", "/sys/menus", "read", "", ""},
	{"team-leader", "*", "/sys/dictionaries", "GET", "", ""},
	{"team-leader", "*", "/sys/dictionaries", "read", "", ""},

	// Organization and content access
	{"team-leader", "*", "/org/groups", "GET", "", ""},
	{"team-leader", "*", "/org/groups", "read", "", ""},
	{"team-leader", "*", "/cms/topics", "GET", "", ""},
	{"team-leader", "*", "/cms/topics", "read", "", ""},
	{"team-leader", "*", "/flow/tasks", "GET", "", ""},
	{"team-leader", "*", "/flow/tasks", "read", "", ""},
	{"team-leader", "*", "/rt/notifications", "GET", "", ""},
	{"team-leader", "*", "/rt/notifications", "read", "", ""},
	{"team-leader", "*", "/res", "GET", "", ""},
	{"team-leader", "*", "/res", "read", "", ""},

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
	{"employee", "*", "/flow/tasks", "GET", "", ""},
	{"employee", "*", "/flow/tasks", "read", "", ""},
	{"employee", "*", "/rt/notifications", "GET", "", ""},
	{"employee", "*", "/rt/notifications", "read", "", ""},
	{"employee", "*", "/res", "GET", "", ""},
	{"employee", "*", "/res", "read", "", ""},

	// Contractor - limited external access
	{"contractor", "*", "/", "GET", "", ""},
	{"contractor", "*", "/", "read", "", ""},
	{"contractor", "*", "/account", "GET", "", ""},
	{"contractor", "*", "/account", "read", "", ""},
	{"contractor", "*", "/sys/menus", "GET", "", ""},
	{"contractor", "*", "/sys/menus", "read", "", ""},
	{"contractor", "*", "/sys/dictionaries", "GET", "", ""},
	{"contractor", "*", "/sys/dictionaries", "read", "", ""},
	{"contractor", "*", "/cms/topics", "GET", "", ""},
	{"contractor", "*", "/cms/topics", "read", "", ""},
	{"contractor", "*", "/res", "GET", "", ""},
	{"contractor", "*", "/res", "read", "", ""},
}

// RoleInheritanceRules for g matcher: child_role, parent_role, domain
var RoleInheritanceRules = [][]string{
	{"system-admin", "super-admin", "*"},
	{"enterprise-admin", "system-admin", "*"},
	{"department-manager", "enterprise-admin", "*"},
	{"team-leader", "department-manager", "*"},
	{"employee", "team-leader", "*"},
}

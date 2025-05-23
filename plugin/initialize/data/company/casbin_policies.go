package company

// CasbinPolicyRules defines policy rules: sub, dom, obj, act, v4, v5
var CasbinPolicyRules = [][]string{
	// Super Admin - full access
	{"super-admin", "*", "*", "*", "", ""},

	// System Admin - system management
	{"system-admin", "*", "/iam/*", "*", "", ""},
	{"system-admin", "*", "/sys/*", "*", "", ""},
	{"system-admin", "*", "/user/*", "*", "", ""},
	{"system-admin", "*", "/access/*", "*", "", ""},
	{"system-admin", "*", "/tenant/*", "*", "", ""},
	{"system-admin", "*", "/space/*", "*", "", ""},
	{"system-admin", "*", "/content/*", "*", "", ""},
	{"system-admin", "*", "/workflow/*", "*", "", ""},

	// Enterprise Admin - business management
	{"enterprise-admin", "*", "/iam/account", "GET", "", ""},
	{"enterprise-admin", "*", "/user/*", "*", "", ""},
	{"enterprise-admin", "*", "/sys/menus", "GET", "", ""},
	{"enterprise-admin", "*", "/sys/dictionaries", "GET", "", ""},
	{"enterprise-admin", "*", "/space/groups", "*", "", ""},
	{"enterprise-admin", "*", "/content/*", "*", "", ""},
	{"enterprise-admin", "*", "/workflow/*", "*", "", ""},

	// Manager - departmental management
	{"manager", "*", "/iam/account", "GET", "", ""},
	{"manager", "*", "/user/users", "GET", "", ""},
	{"manager", "*", "/user/employees", "GET", "", ""},
	{"manager", "*", "/user/employees/*", "PUT", "", ""},
	{"manager", "*", "/sys/menus", "GET", "", ""},
	{"manager", "*", "/sys/dictionaries", "GET", "", ""},
	{"manager", "*", "/space/groups", "GET", "", ""},
	{"manager", "*", "/content/topics", "GET", "", ""},
	{"manager", "*", "/workflow/tasks", "*", "", ""},

	// Employee - basic access
	{"employee", "*", "/iam/account", "GET", "", ""},
	{"employee", "*", "/sys/menus", "GET", "", ""},
	{"employee", "*", "/sys/dictionaries", "GET", "", ""},
	{"employee", "*", "/user/users", "GET", "", ""},
	{"employee", "*", "/user/employees", "GET", "", ""},
	{"employee", "*", "/space/groups", "GET", "", ""},
	{"employee", "*", "/content/topics", "GET", "", ""},
	{"employee", "*", "/workflow/tasks", "GET", "", ""},

	// Guest - read-only access
	{"guest", "*", "/iam/account", "GET", "", ""},
	{"guest", "*", "/sys/menus", "GET", "", ""},
	{"guest", "*", "/content/topics", "GET", "", ""},
}

// RoleInheritanceRules defines role inheritance: child_role, parent_role, domain
var RoleInheritanceRules = [][]string{
	{"system-admin", "super-admin", "*", "", ""},
	{"enterprise-admin", "system-admin", "*", "", ""},
	{"manager", "enterprise-admin", "*", "", ""},
	{"employee", "manager", "*", "", ""},
	{"guest", "employee", "*", "", ""},
}

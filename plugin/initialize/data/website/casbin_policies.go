package website

// CasbinPolicyRules for regular websites
var CasbinPolicyRules = [][]string{
	// Super Admin - full access
	{"super-admin", "*", "*", "*"},

	// Admin - site management
	{"admin", "*", "/", "GET"},
	{"admin", "*", "/swagger/*", "GET"},
	{"admin", "*", "/sys/*", "*"},
	{"admin", "*", "/cms/*", "*"},
	{"admin", "*", "/res/*", "*"},
	{"admin", "*", "/rt/notifications", "*"},
	{"admin", "*", "/account", "GET"},
	{"admin", "*", "/account/space", "GET"},

	// Manager - content management
	{"manager", "*", "/", "GET"},
	{"manager", "*", "/account", "GET"},
	{"manager", "*", "/sys/menus", "GET"},
	{"manager", "*", "/sys/dictionaries", "GET"},
	{"manager", "*", "/sys/users", "GET"},
	{"manager", "*", "/cms/topics", "*"},
	{"manager", "*", "/cms/media", "*"},
	{"manager", "*", "/res", "GET"},
	{"manager", "*", "/rt/notifications", "GET"},

	// Member - standard access
	{"member", "*", "/", "GET"},
	{"member", "*", "/account", "GET"},
	{"member", "*", "/sys/menus", "GET"},
	{"member", "*", "/sys/dictionaries", "GET"},
	{"member", "*", "/cms/topics", "GET"},
	{"member", "*", "/res", "GET"},
	{"member", "*", "/rt/notifications", "GET"},

	// Viewer - read-only
	{"viewer", "*", "/", "GET"},
	{"viewer", "*", "/account", "GET"},
	{"viewer", "*", "/sys/menus", "GET"},
	{"viewer", "*", "/cms/topics", "GET"},
}

// RoleInheritanceRules for websites
var RoleInheritanceRules = [][]string{
	{"admin", "super-admin", "*"},
	{"manager", "admin", "*"},
	{"member", "manager", "*"},
	{"viewer", "member", "*"},
}

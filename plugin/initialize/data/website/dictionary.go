package website

import systemStructs "ncobase/system/structs"

// SystemDefaultDictionaries for regular websites
var SystemDefaultDictionaries = []systemStructs.DictionaryBody{
	// User status
	{
		Name:        "User Status",
		Slug:        "user_status",
		Type:        "object",
		Value:       `{"active":"Active","inactive":"Inactive","pending":"Pending"}`,
		Description: "User account status",
	},

	// User roles
	{
		Name:        "User Role",
		Slug:        "user_role",
		Type:        "object",
		Value:       `{"admin":"Administrator","manager":"Manager","member":"Member","viewer":"Viewer"}`,
		Description: "Website user roles",
	},

	// Content status
	{
		Name:        "Content Status",
		Slug:        "content_status",
		Type:        "object",
		Value:       `{"draft":"Draft","published":"Published","archived":"Archived"}`,
		Description: "Content publication status",
	},

	// Priority levels
	{
		Name:        "Priority",
		Slug:        "priority",
		Type:        "object",
		Value:       `{"low":"Low","medium":"Medium","high":"High"}`,
		Description: "Priority levels",
	},

	// Languages
	{
		Name:        "Language",
		Slug:        "language",
		Type:        "object",
		Value:       `{"en-US":"English","zh-CN":"Chinese","es-ES":"Spanish","fr-FR":"French"}`,
		Description: "Supported languages",
	},

	// Menu types
	{
		Name:        "Menu Type",
		Slug:        "menu_type",
		Type:        "object",
		Value:       `{"header":"Header Menu","sidebar":"Sidebar Menu","account":"Account Menu"}`,
		Description: "Menu types",
	},
}

package company

import systemStructs "ncobase/system/structs"

// SystemDefaultDictionaries defines core system dictionary data
var SystemDefaultDictionaries = []systemStructs.DictionaryBody{
	// Employee status
	{
		Name:        "Employee Status",
		Slug:        "employee_status",
		Type:        "object",
		Value:       `{"active":"Active","inactive":"Inactive","on_leave":"On Leave","terminated":"Terminated"}`,
		Description: "Employee status options",
	},

	// Employment types
	{
		Name:        "Employment Type",
		Slug:        "employment_type",
		Type:        "object",
		Value:       `{"full_time":"Full Time","part_time":"Part Time","contract":"Contract","intern":"Intern"}`,
		Description: "Employment type options",
	},

	// Department types
	{
		Name:        "Department Type",
		Slug:        "department_type",
		Type:        "object",
		Value:       `{"executive":"Executive","hr":"Human Resources","finance":"Finance","technology":"Technology","operations":"Operations"}`,
		Description: "Department categories",
	},

	// User roles
	{
		Name:        "User Role",
		Slug:        "user_role",
		Type:        "object",
		Value:       `{"super-admin":"Super Admin","system-admin":"System Admin","enterprise-admin":"Enterprise Admin","manager":"Manager","employee":"Employee","guest":"Guest"}`,
		Description: "System user roles",
	},

	// User status
	{
		Name:        "User Status",
		Slug:        "user_status",
		Type:        "object",
		Value:       `{"active":"Active","inactive":"Inactive","pending":"Pending","locked":"Locked"}`,
		Description: "User account status",
	},

	// Priority levels
	{
		Name:        "Priority",
		Slug:        "priority",
		Type:        "object",
		Value:       `{"low":"Low","medium":"Medium","high":"High","urgent":"Urgent"}`,
		Description: "Priority levels for tasks and tickets",
	},

	// Task status
	{
		Name:        "Task Status",
		Slug:        "task_status",
		Type:        "object",
		Value:       `{"todo":"To Do","in_progress":"In Progress","review":"Under Review","done":"Done","cancelled":"Cancelled"}`,
		Description: "Task workflow status",
	},

	// Languages
	{
		Name:        "Language",
		Slug:        "language",
		Type:        "object",
		Value:       `{"en-US":"English (US)","zh-CN":"Chinese (Simplified)","ja-JP":"Japanese","ko-KR":"Korean"}`,
		Description: "Supported languages",
	},

	// Currencies
	{
		Name:        "Currency",
		Slug:        "currency",
		Type:        "object",
		Value:       `{"USD":"US Dollar","EUR":"Euro","GBP":"British Pound","JPY":"Japanese Yen","CNY":"Chinese Yuan"}`,
		Description: "Supported currencies",
	},

	// Menu types
	{
		Name:        "Menu Type",
		Slug:        "menu_type",
		Type:        "object",
		Value:       `{"header":"Header Menu","sidebar":"Sidebar Menu","submenu":"Submenu","account":"Account Menu","space":"Space Menu"}`,
		Description: "System menu types",
	},
}

package data

import "ncobase/system/structs"

// SystemDefaultDictionaries defines system default dictionary data
var SystemDefaultDictionaries = []structs.DictionaryBody{
	// User status
	{
		Name:        "User Status",
		Slug:        "user_status",
		Type:        "object",
		Value:       `{"active":"Active","inactive":"Inactive","pending":"Pending","locked":"Locked"}`,
		Description: "System user status enumeration",
	},

	// Gender
	{
		Name:        "Gender",
		Slug:        "gender",
		Type:        "object",
		Value:       `{"male":"Male","female":"Female","other":"Other"}`,
		Description: "Gender options",
	},

	// Document types
	{
		Name:        "Document Type",
		Slug:        "document_type",
		Type:        "object",
		Value:       `{"text":"Text Document","spreadsheet":"Spreadsheet","presentation":"Presentation","image":"Image","video":"Video","audio":"Audio","pdf":"PDF Document","archive":"Archive","code":"Code File","other":"Other"}`,
		Description: "Document types supported by the system",
	},

	// File size units
	{
		Name:        "File Size Unit",
		Slug:        "file_size_unit",
		Type:        "object",
		Value:       `{"B":"Bytes","KB":"Kilobytes","MB":"Megabytes","GB":"Gigabytes","TB":"Terabytes"}`,
		Description: "File size units",
	},

	// Role types
	{
		Name:        "Role Type",
		Slug:        "role_type",
		Type:        "object",
		Value:       `{"system":"System Role","organization":"Organization Role","custom":"Custom Role"}`,
		Description: "System role types",
	},

	// Permission types
	{
		Name:        "Permission Type",
		Slug:        "permission_type",
		Type:        "object",
		Value:       `{"menu":"Menu Permission","operation":"Operation Permission","data":"Data Permission","api":"API Permission"}`,
		Description: "System permission types",
	},

	// Notification types
	{
		Name:        "Notification Type",
		Slug:        "notification_type",
		Type:        "object",
		Value:       `{"system":"System Notification","task":"Task Notification","message":"Message Notification","alert":"Alert Notification"}`,
		Description: "System notification types",
	},

	// Priority levels
	{
		Name:        "Priority",
		Slug:        "priority",
		Type:        "object",
		Value:       `{"low":"Low","medium":"Medium","high":"High","urgent":"Urgent"}`,
		Description: "Task or ticket priority levels",
	},

	// Menu types
	{
		Name:        "Menu Type",
		Slug:        "menu_type",
		Type:        "object",
		Value:       `{"header":"Header Menu","sidebar":"Sidebar Menu","submenu":"Submenu","divider":"Divider","account":"Account Menu","tenant":"Tenant Menu"}`,
		Description: "System menu types",
	},

	// Date formats
	{
		Name:        "Date Format",
		Slug:        "date_format",
		Type:        "object",
		Value:       `{"YYYY-MM-DD":"2023-05-21","DD/MM/YYYY":"21/05/2023","MM/DD/YYYY":"05/21/2023","YYYY/MM/DD":"2023/05/21"}`,
		Description: "Date formats supported by the system",
	},

	// Time formats
	{
		Name:        "Time Format",
		Slug:        "time_format",
		Type:        "object",
		Value:       `{"HH:mm:ss":"24-hour format (13:30:00)","hh:mm:ss a":"12-hour format (01:30:00 PM)"}`,
		Description: "Time formats supported by the system",
	},

	// Languages
	{
		Name:        "Language",
		Slug:        "language",
		Type:        "object",
		Value:       `{"en-US":"English (US)","zh-CN":"Chinese (Simplified)","ja-JP":"Japanese","ko-KR":"Korean","fr-FR":"French","de-DE":"German","es-ES":"Spanish"}`,
		Description: "Languages supported by the system",
	},

	// Currencies
	{
		Name:        "Currency",
		Slug:        "currency",
		Type:        "object",
		Value:       `{"USD":{"name":"US Dollar","symbol":"$"},"EUR":{"name":"Euro","symbol":"€"},"GBP":{"name":"British Pound","symbol":"£"},"JPY":{"name":"Japanese Yen","symbol":"¥"},"CNY":{"name":"Chinese Yuan","symbol":"¥"},"HKD":{"name":"Hong Kong Dollar","symbol":"HK$"}}`,
		Description: "Currencies supported by the system",
	},

	// Countries/Regions
	{
		Name:        "Country/Region",
		Slug:        "country_region",
		Type:        "object",
		Value:       `{"US":"United States","CA":"Canada","GB":"United Kingdom","DE":"Germany","FR":"France","JP":"Japan","CN":"China","AU":"Australia","SG":"Singapore","HK":"Hong Kong, China","TW":"Taiwan, China","KR":"South Korea"}`,
		Description: "Common countries and regions",
	},

	// Workflow status
	{
		Name:        "Workflow Status",
		Slug:        "workflow_status",
		Type:        "object",
		Value:       `{"draft":"Draft","pending":"Pending Approval","approved":"Approved","rejected":"Rejected","in_progress":"In Progress","completed":"Completed","cancelled":"Cancelled"}`,
		Description: "Workflow process status",
	},

	// Project status
	{
		Name:        "Project Status",
		Slug:        "project_status",
		Type:        "object",
		Value:       `{"planning":"Planning","in_progress":"In Progress","on_hold":"On Hold","completed":"Completed","cancelled":"Cancelled"}`,
		Description: "Project status",
	},

	// Task status
	{
		Name:        "Task Status",
		Slug:        "task_status",
		Type:        "object",
		Value:       `{"todo":"To Do","in_progress":"In Progress","under_review":"Under Review","done":"Done","cancelled":"Cancelled"}`,
		Description: "Task status",
	},

	// Data sensitivity
	{
		Name:        "Data Sensitivity",
		Slug:        "data_sensitivity",
		Type:        "object",
		Value:       `{"public":"Public","internal":"Internal","confidential":"Confidential","restricted":"Restricted"}`,
		Description: "Data sensitivity levels",
	},

	// Employee status
	{
		Name:        "Employee Status",
		Slug:        "employee_status",
		Type:        "object",
		Value:       `{"full_time":"Full Time","part_time":"Part Time","contract":"Contract","intern":"Intern","probation":"Probation","resigned":"Resigned","terminated":"Terminated"}`,
		Description: "Employee employment status",
	},

	// Department types
	{
		Name:        "Department Type",
		Slug:        "department_type",
		Type:        "object",
		Value:       `{"administrative":"Administrative","business":"Business","engineering":"Engineering","operations":"Operations","sales":"Sales","marketing":"Marketing","finance":"Finance","hr":"HR","r_and_d":"R&D","support":"Support","other":"Other"}`,
		Description: "Types of organizational departments",
	},

	// Meeting types
	{
		Name:        "Meeting Type",
		Slug:        "meeting_type",
		Type:        "object",
		Value:       `{"team":"Team Meeting","project":"Project Meeting","client":"Client Meeting","strategy":"Strategy Meeting","review":"Review Meeting","training":"Training Session","workshop":"Workshop","other":"Other"}`,
		Description: "Types of meetings",
	},

	// Integration types
	{
		Name:        "Integration Type",
		Slug:        "integration_type",
		Type:        "object",
		Value:       `{"api":"API Integration","webhook":"Webhook","data_sync":"Data Synchronization","sso":"Single Sign-On","embedded":"Embedded Integration","other":"Other"}`,
		Description: "Types of system integrations",
	},
}

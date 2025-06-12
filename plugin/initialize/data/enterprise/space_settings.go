package enterprise

import spaceStructs "ncobase/space/structs"

// SystemDefaultSpaceSettings defines default settings for enterprise spaces
var SystemDefaultSpaceSettings = []spaceStructs.CreateSpaceSettingBody{
	{
		SpaceSettingBody: spaceStructs.SpaceSettingBody{
			SettingKey:   "theme_color",
			SettingName:  "Theme Color",
			SettingValue: "#28a745",
			DefaultValue: "#28a745",
			SettingType:  spaceStructs.TypeString,
			Scope:        spaceStructs.Scope,
			Category:     "appearance",
			Description:  "Primary theme color for enterprise branding",
			IsPublic:     true,
			IsRequired:   false,
			IsReadonly:   false,
		},
	},
	{
		SpaceSettingBody: spaceStructs.SpaceSettingBody{
			SettingKey:   "company_name",
			SettingName:  "Company Name",
			SettingValue: "Digital Enterprise",
			DefaultValue: "",
			SettingType:  spaceStructs.TypeString,
			Scope:        spaceStructs.Scope,
			Category:     "general",
			Description:  "Enterprise display name",
			IsPublic:     true,
			IsRequired:   true,
			IsReadonly:   false,
		},
	},
	{
		SpaceSettingBody: spaceStructs.SpaceSettingBody{
			SettingKey:   "max_file_size",
			SettingName:  "Maximum File Size",
			SettingValue: "200",
			DefaultValue: "200",
			SettingType:  spaceStructs.TypeNumber,
			Scope:        spaceStructs.ScopeSystem,
			Category:     "limits",
			Description:  "Maximum file upload size in MB",
			IsPublic:     false,
			IsRequired:   true,
			IsReadonly:   false,
		},
	},
	{
		SpaceSettingBody: spaceStructs.SpaceSettingBody{
			SettingKey:   "enable_sso",
			SettingName:  "Enable Single Sign-On",
			SettingValue: "true",
			DefaultValue: "false",
			SettingType:  spaceStructs.TypeBoolean,
			Scope:        spaceStructs.ScopeFeature,
			Category:     "security",
			Description:  "Enable SSO authentication for enterprise users",
			IsPublic:     false,
			IsRequired:   false,
			IsReadonly:   false,
		},
	},
	{
		SpaceSettingBody: spaceStructs.SpaceSettingBody{
			SettingKey:   "departments",
			SettingName:  "Departments",
			SettingValue: `["IT", "HR", "Finance", "Marketing", "Operations"]`,
			DefaultValue: `["IT", "HR", "Finance"]`,
			SettingType:  spaceStructs.TypeArray,
			Scope:        spaceStructs.Scope,
			Category:     "organization",
			Description:  "List of company departments",
			IsPublic:     true,
			IsRequired:   false,
			IsReadonly:   false,
		},
	},
	{
		SpaceSettingBody: spaceStructs.SpaceSettingBody{
			SettingKey:   "compliance_settings",
			SettingName:  "Compliance Settings",
			SettingValue: `{"gdpr_enabled": true, "data_retention_days": 2555, "audit_logs": true}`,
			DefaultValue: `{"gdpr_enabled": false, "data_retention_days": 365, "audit_logs": false}`,
			SettingType:  spaceStructs.TypeJSON,
			Scope:        spaceStructs.ScopeSystem,
			Category:     "compliance",
			Description:  "Enterprise compliance and data protection settings",
			IsPublic:     false,
			IsRequired:   true,
			IsReadonly:   false,
		},
	},
}

package company

import spaceStructs "ncobase/core/space/structs"

// SystemDefaultSpaceSettings defines default settings for company spaces
var SystemDefaultSpaceSettings = []spaceStructs.CreateSpaceSettingBody{
	{
		SpaceSettingBody: spaceStructs.SpaceSettingBody{
			SettingKey:   "theme_color",
			SettingName:  "Theme Color",
			SettingValue: "#007bff",
			DefaultValue: "#007bff",
			SettingType:  spaceStructs.TypeString,
			Scope:        spaceStructs.Scope,
			Category:     "appearance",
			Description:  "Primary theme color for company branding",
			IsPublic:     true,
			IsRequired:   false,
			IsReadonly:   false,
		},
	},
	{
		SpaceSettingBody: spaceStructs.SpaceSettingBody{
			SettingKey:   "company_name",
			SettingName:  "Company Name",
			SettingValue: "Digital Company",
			DefaultValue: "",
			SettingType:  spaceStructs.TypeString,
			Scope:        spaceStructs.Scope,
			Category:     "general",
			Description:  "Company display name",
			IsPublic:     true,
			IsRequired:   true,
			IsReadonly:   false,
		},
	},
	{
		SpaceSettingBody: spaceStructs.SpaceSettingBody{
			SettingKey:   "max_file_size",
			SettingName:  "Maximum File Size",
			SettingValue: "50",
			DefaultValue: "50",
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
			SettingKey:   "enable_notifications",
			SettingName:  "Enable Notifications",
			SettingValue: "true",
			DefaultValue: "true",
			SettingType:  spaceStructs.TypeBoolean,
			Scope:        spaceStructs.ScopeFeature,
			Category:     "features",
			Description:  "Enable system notifications",
			IsPublic:     true,
			IsRequired:   false,
			IsReadonly:   false,
		},
	},
	{
		SpaceSettingBody: spaceStructs.SpaceSettingBody{
			SettingKey:   "working_hours",
			SettingName:  "Working Hours",
			SettingValue: `{"start": "09:00", "end": "17:00", "timezone": "UTC"}`,
			DefaultValue: `{"start": "09:00", "end": "17:00", "timezone": "UTC"}`,
			SettingType:  spaceStructs.TypeJSON,
			Scope:        spaceStructs.Scope,
			Category:     "general",
			Description:  "Company working hours configuration",
			IsPublic:     true,
			IsRequired:   false,
			IsReadonly:   false,
		},
	},
}

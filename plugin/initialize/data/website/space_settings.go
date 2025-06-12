package website

import spaceStructs "ncobase/space/structs"

// SystemDefaultSpaceSettings defines default settings for website spaces
var SystemDefaultSpaceSettings = []spaceStructs.CreateSpaceSettingBody{
	{
		SpaceSettingBody: spaceStructs.SpaceSettingBody{
			SettingKey:   "theme_color",
			SettingName:  "Theme Color",
			SettingValue: "#6c757d",
			DefaultValue: "#6c757d",
			SettingType:  spaceStructs.TypeString,
			Scope:        spaceStructs.Scope,
			Category:     "appearance",
			Description:  "Primary theme color for website",
			IsPublic:     true,
			IsRequired:   false,
			IsReadonly:   false,
		},
	},
	{
		SpaceSettingBody: spaceStructs.SpaceSettingBody{
			SettingKey:   "site_name",
			SettingName:  "Site Name",
			SettingValue: "Website Platform",
			DefaultValue: "",
			SettingType:  spaceStructs.TypeString,
			Scope:        spaceStructs.Scope,
			Category:     "general",
			Description:  "Website display name",
			IsPublic:     true,
			IsRequired:   true,
			IsReadonly:   false,
		},
	},
	{
		SpaceSettingBody: spaceStructs.SpaceSettingBody{
			SettingKey:   "max_file_size",
			SettingName:  "Maximum File Size",
			SettingValue: "10",
			DefaultValue: "10",
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
			SettingKey:   "enable_comments",
			SettingName:  "Enable Comments",
			SettingValue: "true",
			DefaultValue: "true",
			SettingType:  spaceStructs.TypeBoolean,
			Scope:        spaceStructs.ScopeFeature,
			Category:     "features",
			Description:  "Enable commenting system on website",
			IsPublic:     true,
			IsRequired:   false,
			IsReadonly:   false,
		},
	},
}

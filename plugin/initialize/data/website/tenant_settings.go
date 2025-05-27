package website

import tenantStructs "ncobase/tenant/structs"

// SystemDefaultTenantSettings defines default settings for website tenants
var SystemDefaultTenantSettings = []tenantStructs.CreateTenantSettingBody{
	{
		TenantSettingBody: tenantStructs.TenantSettingBody{
			SettingKey:   "theme_color",
			SettingName:  "Theme Color",
			SettingValue: "#6c757d",
			DefaultValue: "#6c757d",
			SettingType:  tenantStructs.TypeString,
			Scope:        tenantStructs.ScopeTenant,
			Category:     "appearance",
			Description:  "Primary theme color for website",
			IsPublic:     true,
			IsRequired:   false,
			IsReadonly:   false,
		},
	},
	{
		TenantSettingBody: tenantStructs.TenantSettingBody{
			SettingKey:   "site_name",
			SettingName:  "Site Name",
			SettingValue: "Website Platform",
			DefaultValue: "",
			SettingType:  tenantStructs.TypeString,
			Scope:        tenantStructs.ScopeTenant,
			Category:     "general",
			Description:  "Website display name",
			IsPublic:     true,
			IsRequired:   true,
			IsReadonly:   false,
		},
	},
	{
		TenantSettingBody: tenantStructs.TenantSettingBody{
			SettingKey:   "max_file_size",
			SettingName:  "Maximum File Size",
			SettingValue: "10",
			DefaultValue: "10",
			SettingType:  tenantStructs.TypeNumber,
			Scope:        tenantStructs.ScopeSystem,
			Category:     "limits",
			Description:  "Maximum file upload size in MB",
			IsPublic:     false,
			IsRequired:   true,
			IsReadonly:   false,
		},
	},
	{
		TenantSettingBody: tenantStructs.TenantSettingBody{
			SettingKey:   "enable_comments",
			SettingName:  "Enable Comments",
			SettingValue: "true",
			DefaultValue: "true",
			SettingType:  tenantStructs.TypeBoolean,
			Scope:        tenantStructs.ScopeFeature,
			Category:     "features",
			Description:  "Enable commenting system on website",
			IsPublic:     true,
			IsRequired:   false,
			IsReadonly:   false,
		},
	},
}

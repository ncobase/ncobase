package company

import tenantStructs "ncobase/tenant/structs"

// SystemDefaultTenantSettings defines default settings for company tenants
var SystemDefaultTenantSettings = []tenantStructs.CreateTenantSettingBody{
	{
		TenantSettingBody: tenantStructs.TenantSettingBody{
			SettingKey:   "theme_color",
			SettingName:  "Theme Color",
			SettingValue: "#007bff",
			DefaultValue: "#007bff",
			SettingType:  tenantStructs.TypeString,
			Scope:        tenantStructs.ScopeTenant,
			Category:     "appearance",
			Description:  "Primary theme color for company branding",
			IsPublic:     true,
			IsRequired:   false,
			IsReadonly:   false,
		},
	},
	{
		TenantSettingBody: tenantStructs.TenantSettingBody{
			SettingKey:   "company_name",
			SettingName:  "Company Name",
			SettingValue: "Digital Company",
			DefaultValue: "",
			SettingType:  tenantStructs.TypeString,
			Scope:        tenantStructs.ScopeTenant,
			Category:     "general",
			Description:  "Company display name",
			IsPublic:     true,
			IsRequired:   true,
			IsReadonly:   false,
		},
	},
	{
		TenantSettingBody: tenantStructs.TenantSettingBody{
			SettingKey:   "max_file_size",
			SettingName:  "Maximum File Size",
			SettingValue: "50",
			DefaultValue: "50",
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
			SettingKey:   "enable_notifications",
			SettingName:  "Enable Notifications",
			SettingValue: "true",
			DefaultValue: "true",
			SettingType:  tenantStructs.TypeBoolean,
			Scope:        tenantStructs.ScopeFeature,
			Category:     "features",
			Description:  "Enable system notifications",
			IsPublic:     true,
			IsRequired:   false,
			IsReadonly:   false,
		},
	},
	{
		TenantSettingBody: tenantStructs.TenantSettingBody{
			SettingKey:   "working_hours",
			SettingName:  "Working Hours",
			SettingValue: `{"start": "09:00", "end": "17:00", "timezone": "UTC"}`,
			DefaultValue: `{"start": "09:00", "end": "17:00", "timezone": "UTC"}`,
			SettingType:  tenantStructs.TypeJSON,
			Scope:        tenantStructs.ScopeTenant,
			Category:     "general",
			Description:  "Company working hours configuration",
			IsPublic:     true,
			IsRequired:   false,
			IsReadonly:   false,
		},
	},
}

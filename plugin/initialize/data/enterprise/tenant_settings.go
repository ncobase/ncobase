package enterprise

import tenantStructs "ncobase/tenant/structs"

// SystemDefaultTenantSettings defines default settings for enterprise tenants
var SystemDefaultTenantSettings = []tenantStructs.CreateTenantSettingBody{
	{
		TenantSettingBody: tenantStructs.TenantSettingBody{
			SettingKey:   "theme_color",
			SettingName:  "Theme Color",
			SettingValue: "#28a745",
			DefaultValue: "#28a745",
			SettingType:  tenantStructs.TypeString,
			Scope:        tenantStructs.ScopeTenant,
			Category:     "appearance",
			Description:  "Primary theme color for enterprise branding",
			IsPublic:     true,
			IsRequired:   false,
			IsReadonly:   false,
		},
	},
	{
		TenantSettingBody: tenantStructs.TenantSettingBody{
			SettingKey:   "company_name",
			SettingName:  "Company Name",
			SettingValue: "Digital Enterprise",
			DefaultValue: "",
			SettingType:  tenantStructs.TypeString,
			Scope:        tenantStructs.ScopeTenant,
			Category:     "general",
			Description:  "Enterprise display name",
			IsPublic:     true,
			IsRequired:   true,
			IsReadonly:   false,
		},
	},
	{
		TenantSettingBody: tenantStructs.TenantSettingBody{
			SettingKey:   "max_file_size",
			SettingName:  "Maximum File Size",
			SettingValue: "200",
			DefaultValue: "200",
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
			SettingKey:   "enable_sso",
			SettingName:  "Enable Single Sign-On",
			SettingValue: "true",
			DefaultValue: "false",
			SettingType:  tenantStructs.TypeBoolean,
			Scope:        tenantStructs.ScopeFeature,
			Category:     "security",
			Description:  "Enable SSO authentication for enterprise users",
			IsPublic:     false,
			IsRequired:   false,
			IsReadonly:   false,
		},
	},
	{
		TenantSettingBody: tenantStructs.TenantSettingBody{
			SettingKey:   "departments",
			SettingName:  "Departments",
			SettingValue: `["IT", "HR", "Finance", "Marketing", "Operations"]`,
			DefaultValue: `["IT", "HR", "Finance"]`,
			SettingType:  tenantStructs.TypeArray,
			Scope:        tenantStructs.ScopeTenant,
			Category:     "organization",
			Description:  "List of company departments",
			IsPublic:     true,
			IsRequired:   false,
			IsReadonly:   false,
		},
	},
	{
		TenantSettingBody: tenantStructs.TenantSettingBody{
			SettingKey:   "compliance_settings",
			SettingName:  "Compliance Settings",
			SettingValue: `{"gdpr_enabled": true, "data_retention_days": 2555, "audit_logs": true}`,
			DefaultValue: `{"gdpr_enabled": false, "data_retention_days": 365, "audit_logs": false}`,
			SettingType:  tenantStructs.TypeJSON,
			Scope:        tenantStructs.ScopeSystem,
			Category:     "compliance",
			Description:  "Enterprise compliance and data protection settings",
			IsPublic:     false,
			IsRequired:   true,
			IsReadonly:   false,
		},
	},
}

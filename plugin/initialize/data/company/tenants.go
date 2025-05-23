package company

import tenantsStructs "ncobase/tenant/structs"

// SystemDefaultTenants defines enterprise tenants
var SystemDefaultTenants = []tenantsStructs.CreateTenantBody{
	{
		TenantBody: tenantsStructs.TenantBody{
			Name:        "Digital Enterprise Platform",
			Slug:        "digital-enterprise",
			Type:        "enterprise",
			Title:       "Digital Enterprise Management Platform",
			URL:         "https://enterprise.digital",
			Description: "Multi-tenant digital enterprise management and collaboration platform",
		},
	},
}

package company

import tenantsStructs "ncobase/tenant/structs"

// SystemDefaultTenants defines company tenants
var SystemDefaultTenants = []tenantsStructs.CreateTenantBody{
	{
		TenantBody: tenantsStructs.TenantBody{
			Name:        "Digital Company Platform",
			Slug:        "digital-company",
			Type:        "company",
			Title:       "Digital Company Management Platform",
			URL:         "https://company.digital",
			Description: "Multi-tenant digital company management and collaboration platform",
		},
	},
}

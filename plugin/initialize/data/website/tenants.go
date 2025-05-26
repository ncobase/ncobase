package website

import tenantStructs "ncobase/tenant/structs"

// SystemDefaultTenants for regular websites
var SystemDefaultTenants = []tenantStructs.CreateTenantBody{
	{
		TenantBody: tenantStructs.TenantBody{
			Name:        "Website Platform",
			Slug:        "website-platform",
			Type:        "website",
			Title:       "Website Management Platform",
			URL:         "https://website.com",
			Description: "Simple website management platform",
		},
	},
}

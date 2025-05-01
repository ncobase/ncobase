package data

import "ncobase/core/tenant/structs"

// SystemDefaultTenants defines the default tenants
var SystemDefaultTenants = []structs.CreateTenantBody{
	{
		TenantBody: structs.TenantBody{
			Name:      "NCOBase Corporation",
			Slug:      "ncobase",
			CreatedBy: nil, // Will be set during initialization
		},
	},
}

package data

import "ncobase/tenant/structs"

// SystemDefaultTenants defines enterprise tenants
var SystemDefaultTenants = []structs.CreateTenantBody{
	{
		TenantBody: structs.TenantBody{
			Name:        "Digital Enterprise Platform",
			Slug:        "digital-enterprise",
			Type:        "enterprise",
			Title:       "Digital Enterprise Management Platform",
			URL:         "https://enterprise.digital",
			Description: "Multi-tenant digital enterprise management and collaboration platform",
		},
	},
	{
		TenantBody: structs.TenantBody{
			Name:        "TechCorp Solutions Tenant",
			Slug:        "techcorp-tenant",
			Type:        "subsidiary",
			Title:       "TechCorp Solutions Platform",
			URL:         "https://techcorp.digital",
			Description: "Technology solutions and software development tenant",
		},
	},
	{
		TenantBody: structs.TenantBody{
			Name:        "MediaCorp Digital Tenant",
			Slug:        "mediacorp-tenant",
			Type:        "subsidiary",
			Title:       "MediaCorp Digital Platform",
			URL:         "https://mediacorp.digital",
			Description: "Digital media and content creation tenant",
		},
	},
}

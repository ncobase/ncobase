package website

import tenantStructs "ncobase/tenant/structs"

// SystemDefaultTenantQuotas defines default quotas for website tenants
var SystemDefaultTenantQuotas = []tenantStructs.CreateTenantQuotaBody{
	{
		TenantQuotaBody: tenantStructs.TenantQuotaBody{
			QuotaType:   tenantStructs.QuotaTypeUser,
			QuotaName:   "Maximum Users",
			MaxValue:    10,
			CurrentUsed: 0,
			Unit:        tenantStructs.UnitCount,
			Description: "Maximum number of users allowed in website tenant",
			Enabled:     true,
		},
	},
	{
		TenantQuotaBody: tenantStructs.TenantQuotaBody{
			QuotaType:   tenantStructs.QuotaTypeStorage,
			QuotaName:   "Storage Limit",
			MaxValue:    2,
			CurrentUsed: 0,
			Unit:        tenantStructs.UnitGB,
			Description: "Maximum storage space for website data",
			Enabled:     true,
		},
	},
	{
		TenantQuotaBody: tenantStructs.TenantQuotaBody{
			QuotaType:   tenantStructs.QuotaTypeAPI,
			QuotaName:   "API Calls Per Month",
			MaxValue:    1000,
			CurrentUsed: 0,
			Unit:        tenantStructs.UnitCount,
			Description: "Maximum API calls per month for website operations",
			Enabled:     true,
		},
	},
}

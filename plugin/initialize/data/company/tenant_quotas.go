package company

import tenantStructs "ncobase/tenant/structs"

// SystemDefaultTenantQuotas defines default quotas for company tenants
var SystemDefaultTenantQuotas = []tenantStructs.CreateTenantQuotaBody{
	{
		TenantQuotaBody: tenantStructs.TenantQuotaBody{
			QuotaType:   tenantStructs.QuotaTypeUser,
			QuotaName:   "Maximum Users",
			MaxValue:    100,
			CurrentUsed: 0,
			Unit:        tenantStructs.UnitCount,
			Description: "Maximum number of users allowed in company tenant",
			Enabled:     true,
		},
	},
	{
		TenantQuotaBody: tenantStructs.TenantQuotaBody{
			QuotaType:   tenantStructs.QuotaTypeStorage,
			QuotaName:   "Storage Limit",
			MaxValue:    10,
			CurrentUsed: 0,
			Unit:        tenantStructs.UnitGB,
			Description: "Maximum storage space for company data",
			Enabled:     true,
		},
	},
	{
		TenantQuotaBody: tenantStructs.TenantQuotaBody{
			QuotaType:   tenantStructs.QuotaTypeAPI,
			QuotaName:   "API Calls Per Month",
			MaxValue:    10000,
			CurrentUsed: 0,
			Unit:        tenantStructs.UnitCount,
			Description: "Maximum API calls per month for company operations",
			Enabled:     true,
		},
	},
	{
		TenantQuotaBody: tenantStructs.TenantQuotaBody{
			QuotaType:   tenantStructs.QuotaTypeProject,
			QuotaName:   "Active Projects",
			MaxValue:    25,
			CurrentUsed: 0,
			Unit:        tenantStructs.UnitCount,
			Description: "Maximum number of active projects",
			Enabled:     true,
		},
	},
}

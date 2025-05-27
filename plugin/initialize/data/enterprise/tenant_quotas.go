package enterprise

import tenantStructs "ncobase/tenant/structs"

// SystemDefaultTenantQuotas defines default quotas for enterprise tenants
var SystemDefaultTenantQuotas = []tenantStructs.CreateTenantQuotaBody{
	{
		TenantQuotaBody: tenantStructs.TenantQuotaBody{
			QuotaType:   tenantStructs.QuotaTypeUser,
			QuotaName:   "Maximum Users",
			MaxValue:    500,
			CurrentUsed: 0,
			Unit:        tenantStructs.UnitCount,
			Description: "Maximum number of users allowed in enterprise tenant",
			Enabled:     true,
		},
	},
	{
		TenantQuotaBody: tenantStructs.TenantQuotaBody{
			QuotaType:   tenantStructs.QuotaTypeStorage,
			QuotaName:   "Storage Limit",
			MaxValue:    100,
			CurrentUsed: 0,
			Unit:        tenantStructs.UnitGB,
			Description: "Maximum storage space for enterprise data",
			Enabled:     true,
		},
	},
	{
		TenantQuotaBody: tenantStructs.TenantQuotaBody{
			QuotaType:   tenantStructs.QuotaTypeAPI,
			QuotaName:   "API Calls Per Month",
			MaxValue:    100000,
			CurrentUsed: 0,
			Unit:        tenantStructs.UnitCount,
			Description: "Maximum API calls per month for enterprise operations",
			Enabled:     true,
		},
	},
	{
		TenantQuotaBody: tenantStructs.TenantQuotaBody{
			QuotaType:   tenantStructs.QuotaTypeProject,
			QuotaName:   "Active Projects",
			MaxValue:    100,
			CurrentUsed: 0,
			Unit:        tenantStructs.UnitCount,
			Description: "Maximum number of active projects",
			Enabled:     true,
		},
	},
}

package company

import spaceStructs "ncobase/core/space/structs"

// SystemDefaultSpaceQuotas defines default quotas for company spaces
var SystemDefaultSpaceQuotas = []spaceStructs.CreateSpaceQuotaBody{
	{
		SpaceQuotaBody: spaceStructs.SpaceQuotaBody{
			QuotaType:   spaceStructs.QuotaTypeUser,
			QuotaName:   "Maximum Users",
			MaxValue:    100,
			CurrentUsed: 0,
			Unit:        spaceStructs.UnitCount,
			Description: "Maximum number of users allowed in company space",
			Enabled:     true,
		},
	},
	{
		SpaceQuotaBody: spaceStructs.SpaceQuotaBody{
			QuotaType:   spaceStructs.QuotaTypeStorage,
			QuotaName:   "Storage Limit",
			MaxValue:    10,
			CurrentUsed: 0,
			Unit:        spaceStructs.UnitGB,
			Description: "Maximum storage space for company data",
			Enabled:     true,
		},
	},
	{
		SpaceQuotaBody: spaceStructs.SpaceQuotaBody{
			QuotaType:   spaceStructs.QuotaTypeAPI,
			QuotaName:   "API Calls Per Month",
			MaxValue:    10000,
			CurrentUsed: 0,
			Unit:        spaceStructs.UnitCount,
			Description: "Maximum API calls per month for company operations",
			Enabled:     true,
		},
	},
	{
		SpaceQuotaBody: spaceStructs.SpaceQuotaBody{
			QuotaType:   spaceStructs.QuotaTypeProject,
			QuotaName:   "Active Projects",
			MaxValue:    25,
			CurrentUsed: 0,
			Unit:        spaceStructs.UnitCount,
			Description: "Maximum number of active projects",
			Enabled:     true,
		},
	},
}

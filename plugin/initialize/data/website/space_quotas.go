package website

import spaceStructs "ncobase/core/space/structs"

// SystemDefaultSpaceQuotas defines default quotas for website spaces
var SystemDefaultSpaceQuotas = []spaceStructs.CreateSpaceQuotaBody{
	{
		SpaceQuotaBody: spaceStructs.SpaceQuotaBody{
			QuotaType:   spaceStructs.QuotaTypeUser,
			QuotaName:   "Maximum Users",
			MaxValue:    10,
			CurrentUsed: 0,
			Unit:        spaceStructs.UnitCount,
			Description: "Maximum number of users allowed in website space",
			Enabled:     true,
		},
	},
	{
		SpaceQuotaBody: spaceStructs.SpaceQuotaBody{
			QuotaType:   spaceStructs.QuotaTypeStorage,
			QuotaName:   "Storage Limit",
			MaxValue:    2,
			CurrentUsed: 0,
			Unit:        spaceStructs.UnitGB,
			Description: "Maximum storage space for website data",
			Enabled:     true,
		},
	},
	{
		SpaceQuotaBody: spaceStructs.SpaceQuotaBody{
			QuotaType:   spaceStructs.QuotaTypeAPI,
			QuotaName:   "API Calls Per Month",
			MaxValue:    1000,
			CurrentUsed: 0,
			Unit:        spaceStructs.UnitCount,
			Description: "Maximum API calls per month for website operations",
			Enabled:     true,
		},
	},
}

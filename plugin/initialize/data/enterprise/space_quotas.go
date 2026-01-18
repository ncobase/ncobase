package enterprise

import spaceStructs "ncobase/core/space/structs"

// SystemDefaultSpaceQuotas defines default quotas for enterprise spaces
var SystemDefaultSpaceQuotas = []spaceStructs.CreateSpaceQuotaBody{
	{
		SpaceQuotaBody: spaceStructs.SpaceQuotaBody{
			QuotaType:   spaceStructs.QuotaTypeUser,
			QuotaName:   "Maximum Users",
			MaxValue:    500,
			CurrentUsed: 0,
			Unit:        spaceStructs.UnitCount,
			Description: "Maximum number of users allowed in enterprise space",
			Enabled:     true,
		},
	},
	{
		SpaceQuotaBody: spaceStructs.SpaceQuotaBody{
			QuotaType:   spaceStructs.QuotaTypeStorage,
			QuotaName:   "Storage Limit",
			MaxValue:    100,
			CurrentUsed: 0,
			Unit:        spaceStructs.UnitGB,
			Description: "Maximum storage space for enterprise data",
			Enabled:     true,
		},
	},
	{
		SpaceQuotaBody: spaceStructs.SpaceQuotaBody{
			QuotaType:   spaceStructs.QuotaTypeAPI,
			QuotaName:   "API Calls Per Month",
			MaxValue:    100000,
			CurrentUsed: 0,
			Unit:        spaceStructs.UnitCount,
			Description: "Maximum API calls per month for enterprise operations",
			Enabled:     true,
		},
	},
	{
		SpaceQuotaBody: spaceStructs.SpaceQuotaBody{
			QuotaType:   spaceStructs.QuotaTypeProject,
			QuotaName:   "Active Projects",
			MaxValue:    100,
			CurrentUsed: 0,
			Unit:        spaceStructs.UnitCount,
			Description: "Maximum number of active projects",
			Enabled:     true,
		},
	},
}

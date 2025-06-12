package company

import spacesStructs "ncobase/space/structs"

// SystemDefaultSpaces defines company spaces
var SystemDefaultSpaces = []spacesStructs.CreateSpaceBody{
	{
		SpaceBody: spacesStructs.SpaceBody{
			Name:        "Digital Company Platform",
			Slug:        "digital-company",
			Type:        "company",
			Title:       "Digital Company Management Platform",
			URL:         "https://company.digital",
			Description: "Multi-space digital company management and collaboration platform",
		},
	},
}

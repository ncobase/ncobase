package website

import spaceStructs "ncobase/space/structs"

// SystemDefaultSpaces for regular websites
var SystemDefaultSpaces = []spaceStructs.CreateSpaceBody{
	{
		SpaceBody: spaceStructs.SpaceBody{
			Name:        "Website Platform",
			Slug:        "website-platform",
			Type:        "website",
			Title:       "Website Management Platform",
			URL:         "https://website.com",
			Description: "Simple website management platform",
		},
	},
}

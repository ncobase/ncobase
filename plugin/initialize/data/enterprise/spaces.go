package enterprise

import "ncobase/space/structs"

// SystemDefaultSpaces defines enterprise spaces
var SystemDefaultSpaces = []structs.CreateSpaceBody{
	{
		SpaceBody: structs.SpaceBody{
			Name:        "Digital Enterprise Platform",
			Slug:        "digital-enterprise",
			Type:        "enterprise",
			Title:       "Digital Enterprise Management Platform",
			URL:         "https://enterprise.digital",
			Description: "Multi-space digital enterprise management and collaboration platform",
		},
	},
	{
		SpaceBody: structs.SpaceBody{
			Name:        "TechCorp Solutions Space",
			Slug:        "techcorp-space",
			Type:        "subsidiary",
			Title:       "TechCorp Solutions Platform",
			URL:         "https://techcorp.digital",
			Description: "Technology solutions and software development space",
		},
	},
	{
		SpaceBody: structs.SpaceBody{
			Name:        "MediaCorp Digital Space",
			Slug:        "mediacorp-space",
			Type:        "subsidiary",
			Title:       "MediaCorp Digital Platform",
			URL:         "https://mediacorp.digital",
			Description: "Digital media and content creation space",
		},
	},
}

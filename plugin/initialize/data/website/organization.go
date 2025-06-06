package website

import (
	"ncobase/space/structs"
)

// OrganizationStructure for simple websites
var OrganizationStructure = struct {
	MainGroup structs.GroupBody `json:"main_group"`
}{
	MainGroup: structs.GroupBody{
		Name:        "Website",
		Slug:        "website-platform",
		Type:        "website",
		Description: "Main website organization",
	},
}

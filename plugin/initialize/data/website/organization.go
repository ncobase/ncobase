package website

import (
	"ncobase/core/organization/structs"
)

// OrganizationStructure for simple websites
var OrganizationStructure = struct {
	MainOrganization structs.OrganizationBody `json:"main_organization"`
}{
	MainOrganization: structs.OrganizationBody{
		Name:        "Website",
		Slug:        "website-platform",
		Type:        "website",
		Description: "Main website organization",
	},
}

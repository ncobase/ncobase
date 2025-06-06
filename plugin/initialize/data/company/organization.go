package company

import (
	accessStructs "ncobase/access/structs"
	"ncobase/space/structs"
)

// OrganizationStructure defines company organizational hierarchy
var OrganizationStructure = struct {
	Company           structs.GroupBody           `json:"company"`
	Headquarters      []structs.GroupBody         `json:"headquarters"`
	Subsidiaries      []structs.GroupBody         `json:"subsidiaries"`
	CompanyStructures map[string]CompanyStructure `json:"company_structures"`
	SharedDepartments []Department                `json:"shared_departments"`
	OrganizationRoles []OrganizationRole          `json:"organization_roles"`
}{
	// Main company group
	Company: structs.GroupBody{
		Name:        "Digital Company Group",
		Slug:        "digital-company",
		Type:        "company",
		Description: "Main company organization",
	},

	// Headquarters departments
	Headquarters: []structs.GroupBody{
		{
			Name:        "Executive Office",
			Slug:        "executive-office",
			Type:        "department",
			Description: "Executive leadership and management",
		},
		{
			Name:        "Corporate HR",
			Slug:        "corporate-hr",
			Description: "Company HR management",
		},
		{
			Name:        "Corporate Finance",
			Slug:        "corporate-finance",
			Type:        "department",
			Description: "Company financial management",
		},
	},

	// Subsidiaries under company
	Subsidiaries: []structs.GroupBody{
		{
			Name:        "TechCorp Solutions",
			Slug:        "techcorp",
			Type:        "subsidiary",
			Description: "Technology and software development",
		},
		{
			Name:        "BusinessCorp Services",
			Slug:        "businesscorp",
			Type:        "subsidiary",
			Description: "Business services and consulting",
		},
	},

	// Company-specific structures
	CompanyStructures: map[string]CompanyStructure{
		"techcorp": {
			Departments: []Department{
				{
					Info: structs.GroupBody{
						Name:        "Technology Department",
						Slug:        "technology",
						Type:        "department",
						Description: "Software development and IT",
					},
					Teams: []structs.GroupBody{
						{Name: "Development Team", Slug: "dev-team", Type: "team", Description: "Software development"},
						{Name: "QA Team", Slug: "qa-team", Type: "team", Description: "Quality assurance"},
					},
				},
			},
		},
		"businesscorp": {
			Departments: []Department{
				{
					Info: structs.GroupBody{
						Name: "Business Operations",
						Slug: "business-ops",

						Description: "Business operations and support",
					},
					Teams: []structs.GroupBody{
						{Name: "Operations Team", Slug: "ops-team", Type: "team", Description: "Business operations"},
						{Name: "Support Team", Slug: "support-team", Type: "team", Description: "Customer support"},
					},
				},
			},
		},
	},

	// Shared departments across subsidiaries
	SharedDepartments: []Department{
		{
			Info: structs.GroupBody{
				Name:        "Human Resources",
				Slug:        "%s-hr",
				Type:        "department",
				Description: "Human resources management",
			},
			Teams: []structs.GroupBody{
				{Name: "Recruitment", Slug: "%s-recruitment", Type: "team", Description: "Talent acquisition"},
				{Name: "Employee Relations", Slug: "%s-employee-relations", Type: "team", Description: "Employee support"},
			},
		},
		{
			Info: structs.GroupBody{
				Name:        "Finance & Accounting",
				Slug:        "%s-finance",
				Type:        "department",
				Description: "Financial management",
			},
			Teams: []structs.GroupBody{
				{Name: "Accounting", Slug: "%s-accounting", Type: "team", Description: "Financial accounting"},
				{Name: "Budget Planning", Slug: "%s-budget", Type: "team", Description: "Budget and planning"},
			},
		},
	},

	// Organization-specific roles
	OrganizationRoles: []OrganizationRole{
		{
			Role: accessStructs.RoleBody{
				Name:        "Team Lead",
				Slug:        "team-lead",
				Description: "Team leadership role",
			},
		},
	},
}

// CompanyStructure represents a company's organizational structure
type CompanyStructure struct {
	Departments []Department `json:"departments"`
}

// Department represents a department and its teams
type Department struct {
	Info  structs.GroupBody   `json:"info"`
	Teams []structs.GroupBody `json:"teams"`
}

// OrganizationRole represents organizational roles
type OrganizationRole struct {
	Role accessStructs.RoleBody `json:"role"`
}

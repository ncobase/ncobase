package data

import (
	accessStructs "ncobase/core/access/structs"
	"ncobase/core/space/structs"
)

// OrganizationStructure defines the company hierarchical structure
var OrganizationStructure = struct {
	MainGroup         structs.GroupBody           `json:"main_group"`
	GroupDepartments  []structs.GroupBody         `json:"group_departments"`
	Companies         []structs.GroupBody         `json:"companies"`
	CompanyStructures map[string]CompanyStructure `json:"company_structures"`
	TemporaryGroups   []structs.GroupBody         `json:"temporary_groups"`
	OrganizationRoles []OrganizationRole          `json:"organization_roles"`
}{
	MainGroup: structs.GroupBody{
		Name:      "Enterprise Group",
		Slug:      "enterprise-group",
		TenantID:  nil, // Will be set during initialization
		CreatedBy: nil, // Will be set during initialization
		UpdatedBy: nil, // Will be set during initialization
	},
	GroupDepartments: []structs.GroupBody{
		{
			Name:      "Executive Office",
			Slug:      "executive-office",
			TenantID:  nil, // Will be set during initialization
			CreatedBy: nil, // Will be set during initialization
			UpdatedBy: nil, // Will be set during initialization
		},
		{
			Name:      "Group HR Department",
			Slug:      "group-hr-department",
			TenantID:  nil, // Will be set during initialization
			CreatedBy: nil, // Will be set during initialization
			UpdatedBy: nil, // Will be set during initialization
		},
		{
			Name:      "Group Finance Department",
			Slug:      "group-finance-department",
			TenantID:  nil, // Will be set during initialization
			CreatedBy: nil, // Will be set during initialization
			UpdatedBy: nil, // Will be set during initialization
		},
	},
	Companies: []structs.GroupBody{
		{
			Name:      "Technology Company",
			Slug:      "tech-company",
			TenantID:  nil, // Will be set during initialization
			CreatedBy: nil, // Will be set during initialization
			UpdatedBy: nil, // Will be set during initialization
		},
		{
			Name:      "Media Company",
			Slug:      "media-company",
			TenantID:  nil, // Will be set during initialization
			CreatedBy: nil, // Will be set during initialization
			UpdatedBy: nil, // Will be set during initialization
		},
	},
	CompanyStructures: map[string]CompanyStructure{
		"tech-company": {
			Departments: []Department{
				{
					Info: structs.GroupBody{
						Name:      "Technology Department",
						Slug:      "tech-department",
						TenantID:  nil, // Will be set during initialization
						CreatedBy: nil, // Will be set during initialization
						UpdatedBy: nil, // Will be set during initialization
					},
					Teams: []structs.GroupBody{
						{Name: "Development Team", Slug: "rd-team"},
						{Name: "Operations Team", Slug: "operations-team"},
						{Name: "QA Team", Slug: "qa-team"},
					},
				},
				{
					Info: structs.GroupBody{
						Name:      "Product Department",
						Slug:      "product-department",
						TenantID:  nil, // Will be set during initialization
						CreatedBy: nil, // Will be set during initialization
						UpdatedBy: nil, // Will be set during initialization
					},
					Teams: []structs.GroupBody{
						{Name: "Product Planning Team", Slug: "product-planning-team"},
						{Name: "UX Team", Slug: "ux-team"},
					},
				},
			},
		},
		"media-company": {
			Departments: []Department{
				{
					Info: structs.GroupBody{
						Name:      "Content Production Department",
						Slug:      "content-production-department",
						TenantID:  nil, // Will be set during initialization
						CreatedBy: nil, // Will be set during initialization
						UpdatedBy: nil, // Will be set during initialization
					},
					Teams: []structs.GroupBody{
						{Name: "Video Production Team", Slug: "video-production-team"},
						{Name: "Text Editing Team", Slug: "text-editing-team"},
					},
				},
				{
					Info: structs.GroupBody{
						Name:      "Media Operations Department",
						Slug:      "media-operations-department",
						TenantID:  nil, // Will be set during initialization
						CreatedBy: nil, // Will be set during initialization
						UpdatedBy: nil, // Will be set during initialization
					},
					Teams: []structs.GroupBody{
						{Name: "Social Media Team", Slug: "social-media-team"},
						{Name: "Data Analysis Team", Slug: "data-analysis-team"},
					},
				},
			},
		},
	},
	TemporaryGroups: []structs.GroupBody{
		{
			Name:      "Strategy Committee",
			Slug:      "strategy-committee",
			TenantID:  nil, // Will be set during initialization
			CreatedBy: nil, // Will be set during initialization
			UpdatedBy: nil, // Will be set during initialization
		},
		{
			Name:      "Digital Transformation Team",
			Slug:      "digital-transformation-team",
			TenantID:  nil, // Will be set during initialization
			CreatedBy: nil, // Will be set during initialization
			UpdatedBy: nil, // Will be set during initialization
		},
	},
	OrganizationRoles: []OrganizationRole{
		{
			Role: accessStructs.RoleBody{
				Name: "Group Admin",
				Slug: "group-admin",
			},
			Permissions: []string{"Manage Group", "Manage Department", "Manage Team"},
		},
		{
			Role: accessStructs.RoleBody{
				Name: "Department Manager",
				Slug: "department-manager",
			},
			Permissions: []string{"Manage Department", "Manage Team", "View Group"},
		},
		{
			Role: accessStructs.RoleBody{
				Name: "Team Leader",
				Slug: "team-leader",
			},
			Permissions: []string{"Manage Team", "View Department", "View Group"},
		},
		{
			Role: accessStructs.RoleBody{
				Name: "Employee",
				Slug: "employee",
			},
			Permissions: []string{"View Team", "View Department", "View Group"},
		},
	},
}

// CompanyStructure represents a company's departments and teams
type CompanyStructure struct {
	Departments []Department `json:"departments"`
}

// Department represents a department and its teams
type Department struct {
	Info  structs.GroupBody   `json:"info"`
	Teams []structs.GroupBody `json:"teams"`
}

// OrganizationRole represents a role and its permissions within the organization
type OrganizationRole struct {
	Role        accessStructs.RoleBody `json:"role"`
	Permissions []string               `json:"permissions"`
}

// OrganizationPermissions defines specific permissions for organizational roles
var OrganizationPermissions = []accessStructs.PermissionBody{
	{Name: "Manage Group", Action: "*", Subject: "group"},
	{Name: "Manage Department", Action: "*", Subject: "department"},
	{Name: "Manage Team", Action: "*", Subject: "team"},
	{Name: "View Group", Action: "GET", Subject: "group"},
	{Name: "View Department", Action: "GET", Subject: "department"},
	{Name: "View Team", Action: "GET", Subject: "team"},
}

// CommonDepartments defines departments that exist in all companies
var CommonDepartments = []Department{
	{
		Info: structs.GroupBody{
			Name:      "Marketing Department",
			Slug:      "%s-marketing-department", // Format with company slug
			TenantID:  nil,                       // Will be set during initialization
			CreatedBy: nil,                       // Will be set during initialization
			UpdatedBy: nil,                       // Will be set during initialization
		},
		Teams: []structs.GroupBody{
			{Name: "Brand Team", Slug: "%s-brand-team"},                        // Format with company slug
			{Name: "Market Research Team", Slug: "%s-marketing-research-team"}, // Format with company slug
		},
	},
	{
		Info: structs.GroupBody{
			Name:      "HR Department",
			Slug:      "%s-hr-department", // Format with company slug
			TenantID:  nil,                // Will be set during initialization
			CreatedBy: nil,                // Will be set during initialization
			UpdatedBy: nil,                // Will be set during initialization
		},
		Teams: []structs.GroupBody{
			{Name: "Recruitment Team", Slug: "%s-recruitment-team"},                  // Format with company slug
			{Name: "Training & Development Team", Slug: "%s-train-development-team"}, // Format with company slug
		},
	},
}

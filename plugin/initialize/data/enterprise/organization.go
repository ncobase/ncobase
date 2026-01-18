package enterprise

import (
	accessStructs "ncobase/core/access/structs"
	"ncobase/core/organization/structs"
)

// OrganizationStructure defines the enterprise organizational hierarchy
var OrganizationStructure = struct {
	Enterprise        structs.OrganizationBody    `json:"enterprise"`
	Headquarters      []structs.OrganizationBody  `json:"headquarters"`
	Companies         []structs.OrganizationBody  `json:"companies"`
	CompanyStructures map[string]CompanyStructure `json:"company_structures"`
	SharedDepartments []Department                `json:"shared_departments"`
	OrganizationRoles []OrganizationRole          `json:"organization_roles"`
}{
	Enterprise: structs.OrganizationBody{
		Name:        "Digital Enterprise Group",
		Slug:        "digital-enterprise",
		Type:        "enterprise",
		Description: "Multi-space digital enterprise management platform",
	},
	Headquarters: []structs.OrganizationBody{
		{
			Name:        "Executive Office",
			Slug:        "executive-office",
			Type:        "department",
			Description: "Executive leadership and strategic management",
		},
		{
			Name:        "Corporate HR",
			Slug:        "corporate-hr",
			Type:        "department",
			Description: "Enterprise-wide human resources management",
		},
		{
			Name:        "Corporate Finance",
			Slug:        "corporate-finance",
			Description: "Enterprise financial management and control",
		},
		{
			Name:        "Corporate IT",
			Slug:        "corporate-it",
			Type:        "department",
			Description: "Enterprise IT infrastructure and services",
		},
	},
	Companies: []structs.OrganizationBody{
		{
			Name:        "TechCorp Solutions",
			Slug:        "techcorp",
			Type:        "subsidiary",
			Description: "Technology solutions and software development",
		},
		{
			Name:        "MediaCorp Digital",
			Slug:        "mediacorp",
			Type:        "subsidiary",
			Description: "Digital media and content creation services",
		},
		{
			Name:        "ConsultCorp Advisory",
			Slug:        "consultcorp",
			Type:        "subsidiary",
			Description: "Business consulting and advisory services",
		},
	},
	CompanyStructures: map[string]CompanyStructure{
		"techcorp": {
			Departments: []Department{
				{
					Info: structs.OrganizationBody{
						Name:        "Technology Department",
						Slug:        "technology",
						Type:        "department",
						Description: "Software development and technical operations",
					},
					Teams: []structs.OrganizationBody{
						{Name: "Backend Development", Slug: "backend-dev", Type: "team", Description: "Server-side development team"},
						{Name: "Frontend Development", Slug: "frontend-dev", Type: "team", Description: "Client-side development team"},
						{Name: "DevOps", Slug: "devops", Type: "team", Description: "Development operations and infrastructure"},
						{Name: "QA Engineering", Slug: "qa-engineering", Type: "team", Description: "Quality assurance and testing"},
					},
				},
				{
					Info: structs.OrganizationBody{
						Name:        "Product Management",
						Slug:        "product-management",
						Type:        "department",
						Description: "Product strategy and management",
					},
					Teams: []structs.OrganizationBody{
						{Name: "Product Strategy", Slug: "product-strategy", Type: "team", Description: "Product planning and roadmap"},
						{Name: "UX/UI Design", Slug: "ux-ui-design", Type: "team", Description: "User experience and interface design"},
					},
				},
			},
		},
		"mediacorp": {
			Departments: []Department{
				{
					Info: structs.OrganizationBody{
						Name:        "Content Production",
						Slug:        "content-production",
						Type:        "department",
						Description: "Digital content creation and production",
					},
					Teams: []structs.OrganizationBody{
						{Name: "Video Production", Slug: "video-production", Type: "team", Description: "Video content creation"},
						{Name: "Editorial", Slug: "editorial", Type: "team", Description: "Content writing and editing"},
						{Name: "Graphic Design", Slug: "graphic-design", Type: "team", Description: "Visual design and graphics"},
					},
				},
				{
					Info: structs.OrganizationBody{
						Name:        "Digital Marketing",
						Slug:        "digital-marketing",
						Type:        "department",
						Description: "Digital marketing and promotion",
					},
					Teams: []structs.OrganizationBody{
						{Name: "Social Media", Slug: "social-media", Type: "team", Description: "Social media management"},
						{Name: "SEO/SEM", Slug: "seo-sem", Type: "team", Description: "Search engine optimization and marketing"},
					},
				},
			},
		},
		"consultcorp": {
			Departments: []Department{
				{
					Info: structs.OrganizationBody{
						Name:        "Business Consulting",
						Slug:        "business-consulting",
						Type:        "department",
						Description: "Strategic business consulting services",
					},
					Teams: []structs.OrganizationBody{
						{Name: "Strategy Consulting", Slug: "strategy-consulting", Type: "team", Description: "Strategic planning and advisory"},
						{Name: "Process Optimization", Slug: "process-optimization", Type: "team", Description: "Business process improvement"},
					},
				},
			},
		},
	},
	SharedDepartments: []Department{
		{
			Info: structs.OrganizationBody{
				Name:        "Human Resources",
				Slug:        "%s-hr", // Format with company slug
				Type:        "department",
				Description: "Human resources management",
			},
			Teams: []structs.OrganizationBody{
				{Name: "Recruitment", Slug: "%s-recruitment", Type: "team", Description: "Talent acquisition"},
				{Name: "Employee Relations", Slug: "%s-employee-relations", Type: "team", Description: "Employee support and relations"},
			},
		},
		{
			Info: structs.OrganizationBody{
				Name:        "Finance & Accounting",
				Slug:        "%s-finance",
				Type:        "department",
				Description: "Financial management and accounting",
			},
			Teams: []structs.OrganizationBody{
				{Name: "Accounting", Slug: "%s-accounting", Type: "team", Description: "Financial accounting"},
				{Name: "Financial Planning", Slug: "%s-financial-planning", Type: "team", Description: "Budget and planning"},
			},
		},
		{
			Info: structs.OrganizationBody{
				Name:        "Operations",
				Slug:        "%s-operations",
				Type:        "department",
				Description: "Operational management and support",
			},
			Teams: []structs.OrganizationBody{
				{Name: "Administration", Slug: "%s-administration", Type: "team", Description: "Administrative support"},
				{Name: "Facilities", Slug: "%s-facilities", Type: "team", Description: "Facilities management"},
			},
		},
	},
	OrganizationRoles: []OrganizationRole{
		{
			Role: accessStructs.RoleBody{
				Name:        "Enterprise Executive",
				Slug:        "enterprise-executive",
				Description: "Enterprise-level executive leadership",
			},
			Permissions: []string{
				"System Management",
				"Organization Management",
				"Financial Management",
				"HR Management",
			},
		},
		{
			Role: accessStructs.RoleBody{
				Name:        "Company Director",
				Slug:        "company-director",
				Description: "Company-level leadership and management",
			},
			Permissions: []string{
				"Organization Management",
				"Department Management",
				"Employee Management",
			},
		},
		{
			Role: accessStructs.RoleBody{
				Name:        "Department Head",
				Slug:        "department-head",
				Description: "Department leadership and oversight",
			},
			Permissions: []string{
				"Department Management",
				"Team Management",
				"Employee Management",
			},
		},
		{
			Role: accessStructs.RoleBody{
				Name:        "Team Supervisor",
				Slug:        "team-supervisor",
				Description: "Team supervision and coordination",
			},
			Permissions: []string{
				"Team Management",
				"View Employees",
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
	Info  structs.OrganizationBody   `json:"info"`
	Teams []structs.OrganizationBody `json:"teams"`
}

// OrganizationRole represents organizational roles and permissions
type OrganizationRole struct {
	Role        accessStructs.RoleBody `json:"role"`
	Permissions []string               `json:"permissions"`
}

package data

import (
	accessStructs "ncobase/access/structs"
	"ncobase/space/structs"
)

// EnterpriseOrganizationStructure defines the enterprise organizational hierarchy
var EnterpriseOrganizationStructure = struct {
	Enterprise        structs.GroupBody           `json:"enterprise"`
	Headquarters      []structs.GroupBody         `json:"headquarters"`
	Companies         []structs.GroupBody         `json:"companies"`
	CompanyStructures map[string]CompanyStructure `json:"company_structures"`
	SharedDepartments []Department                `json:"shared_departments"`
	OrganizationRoles []OrganizationRole          `json:"organization_roles"`
}{
	Enterprise: structs.GroupBody{
		Name:        "Digital Enterprise Group",
		Slug:        "digital-enterprise",
		Description: "Multi-tenant digital enterprise management platform",
	},
	Headquarters: []structs.GroupBody{
		{
			Name:        "Executive Office",
			Slug:        "executive-office",
			Description: "Executive leadership and strategic management",
		},
		{
			Name:        "Corporate HR",
			Slug:        "corporate-hr",
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
			Description: "Enterprise IT infrastructure and services",
		},
	},
	Companies: []structs.GroupBody{
		{
			Name:        "TechCorp Solutions",
			Slug:        "techcorp",
			Description: "Technology solutions and software development",
		},
		{
			Name:        "MediaCorp Digital",
			Slug:        "mediacorp",
			Description: "Digital media and content creation services",
		},
		{
			Name:        "ConsultCorp Advisory",
			Slug:        "consultcorp",
			Description: "Business consulting and advisory services",
		},
	},
	CompanyStructures: map[string]CompanyStructure{
		"techcorp": {
			Departments: []Department{
				{
					Info: structs.GroupBody{
						Name:        "Technology Department",
						Slug:        "technology",
						Description: "Software development and technical operations",
					},
					Teams: []structs.GroupBody{
						{Name: "Backend Development", Slug: "backend-dev", Description: "Server-side development team"},
						{Name: "Frontend Development", Slug: "frontend-dev", Description: "Client-side development team"},
						{Name: "DevOps", Slug: "devops", Description: "Development operations and infrastructure"},
						{Name: "QA Engineering", Slug: "qa-engineering", Description: "Quality assurance and testing"},
					},
				},
				{
					Info: structs.GroupBody{
						Name:        "Product Management",
						Slug:        "product-management",
						Description: "Product strategy and management",
					},
					Teams: []structs.GroupBody{
						{Name: "Product Strategy", Slug: "product-strategy", Description: "Product planning and roadmap"},
						{Name: "UX/UI Design", Slug: "ux-ui-design", Description: "User experience and interface design"},
					},
				},
			},
		},
		"mediacorp": {
			Departments: []Department{
				{
					Info: structs.GroupBody{
						Name:        "Content Production",
						Slug:        "content-production",
						Description: "Digital content creation and production",
					},
					Teams: []structs.GroupBody{
						{Name: "Video Production", Slug: "video-production", Description: "Video content creation"},
						{Name: "Editorial", Slug: "editorial", Description: "Content writing and editing"},
						{Name: "Graphic Design", Slug: "graphic-design", Description: "Visual design and graphics"},
					},
				},
				{
					Info: structs.GroupBody{
						Name:        "Digital Marketing",
						Slug:        "digital-marketing",
						Description: "Digital marketing and promotion",
					},
					Teams: []structs.GroupBody{
						{Name: "Social Media", Slug: "social-media", Description: "Social media management"},
						{Name: "SEO/SEM", Slug: "seo-sem", Description: "Search engine optimization and marketing"},
					},
				},
			},
		},
		"consultcorp": {
			Departments: []Department{
				{
					Info: structs.GroupBody{
						Name:        "Business Consulting",
						Slug:        "business-consulting",
						Description: "Strategic business consulting services",
					},
					Teams: []structs.GroupBody{
						{Name: "Strategy Consulting", Slug: "strategy-consulting", Description: "Strategic planning and advisory"},
						{Name: "Process Optimization", Slug: "process-optimization", Description: "Business process improvement"},
					},
				},
			},
		},
	},
	SharedDepartments: []Department{
		{
			Info: structs.GroupBody{
				Name:        "Human Resources",
				Slug:        "%s-hr", // Format with company slug
				Description: "Human resources management",
			},
			Teams: []structs.GroupBody{
				{Name: "Recruitment", Slug: "%s-recruitment", Description: "Talent acquisition"},
				{Name: "Employee Relations", Slug: "%s-employee-relations", Description: "Employee support and relations"},
			},
		},
		{
			Info: structs.GroupBody{
				Name:        "Finance & Accounting",
				Slug:        "%s-finance",
				Description: "Financial management and accounting",
			},
			Teams: []structs.GroupBody{
				{Name: "Accounting", Slug: "%s-accounting", Description: "Financial accounting"},
				{Name: "Financial Planning", Slug: "%s-financial-planning", Description: "Budget and planning"},
			},
		},
		{
			Info: structs.GroupBody{
				Name:        "Operations",
				Slug:        "%s-operations",
				Description: "Operational management and support",
			},
			Teams: []structs.GroupBody{
				{Name: "Administration", Slug: "%s-administration", Description: "Administrative support"},
				{Name: "Facilities", Slug: "%s-facilities", Description: "Facilities management"},
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
	Info  structs.GroupBody   `json:"info"`
	Teams []structs.GroupBody `json:"teams"`
}

// OrganizationRole represents organizational roles and permissions
type OrganizationRole struct {
	Role        accessStructs.RoleBody `json:"role"`
	Permissions []string               `json:"permissions"`
}

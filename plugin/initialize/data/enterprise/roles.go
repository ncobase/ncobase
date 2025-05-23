package enterprise

import accessStructs "ncobase/access/structs"

// SystemDefaultRoles defines enterprise system roles with clear hierarchy
var SystemDefaultRoles = []accessStructs.CreateRoleBody{
	// ========== System Level Roles ==========
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Super Administrator",
			Slug:        "super-admin",
			Disabled:    false,
			Description: "Super administrator with unrestricted system access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "System Administrator",
			Slug:        "system-admin",
			Disabled:    false,
			Description: "System administrator with full platform management capabilities",
		},
	},

	// ========== Enterprise Level Roles ==========
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Enterprise Administrator",
			Slug:        "enterprise-admin",
			Disabled:    false,
			Description: "Enterprise-wide administrative access across all subsidiaries",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Tenant Administrator",
			Slug:        "tenant-admin",
			Disabled:    false,
			Description: "Tenant-level administrative access within specific tenant",
		},
	},

	// ========== Functional Department Roles ==========
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "HR Manager",
			Slug:        "hr-manager",
			Disabled:    false,
			Description: "Human resources management with employee lifecycle control",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Finance Manager",
			Slug:        "finance-manager",
			Disabled:    false,
			Description: "Financial management with budget and payment oversight",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "IT Manager",
			Slug:        "it-manager",
			Disabled:    false,
			Description: "Information technology management and system operations",
		},
	},

	// ========== Management Level Roles ==========
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Department Manager",
			Slug:        "department-manager",
			Disabled:    false,
			Description: "Department-level management with team oversight responsibilities",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Team Leader",
			Slug:        "team-leader",
			Disabled:    false,
			Description: "Team leadership with direct team member management",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Project Manager",
			Slug:        "project-manager",
			Disabled:    false,
			Description: "Project management with workflow and task coordination",
		},
	},

	// ========== Basic Employee Roles ==========
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Senior Employee",
			Slug:        "senior-employee",
			Disabled:    false,
			Description: "Senior employee with extended access and mentoring responsibilities",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Employee",
			Slug:        "employee",
			Disabled:    false,
			Description: "Standard employee with basic platform access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Junior Employee",
			Slug:        "junior-employee",
			Disabled:    false,
			Description: "Junior employee with supervised access",
		},
	},

	// ========== Special Roles ==========
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Contractor",
			Slug:        "contractor",
			Disabled:    false,
			Description: "External contractor with limited project-based access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Consultant",
			Slug:        "consultant",
			Disabled:    false,
			Description: "External consultant with specialized access permissions",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Intern",
			Slug:        "intern",
			Disabled:    false,
			Description: "Intern with supervised learning-focused access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Auditor",
			Slug:        "auditor",
			Disabled:    false,
			Description: "External auditor with read-only compliance access",
		},
	},

	// ========== Professional Technical Roles ==========
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Technical Lead",
			Slug:        "technical-lead",
			Disabled:    false,
			Description: "Technical leadership with system architecture responsibilities",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Developer",
			Slug:        "developer",
			Disabled:    false,
			Description: "Software developer with development environment access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Designer",
			Slug:        "designer",
			Disabled:    false,
			Description: "UI/UX designer with design system access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Content Creator",
			Slug:        "content-creator",
			Disabled:    false,
			Description: "Content creation specialist with media management access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Marketing Specialist",
			Slug:        "marketing-specialist",
			Disabled:    false,
			Description: "Marketing professional with campaign management access",
		},
	},

	// ========== Customer Service Roles ==========
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Customer Service Manager",
			Slug:        "customer-service-manager",
			Disabled:    false,
			Description: "Customer service management with full support access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Customer Service Representative",
			Slug:        "customer-service-rep",
			Disabled:    false,
			Description: "Front-line customer service with ticket management access",
		},
	},

	// ========== Data and Analytics Roles ==========
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Data Analyst",
			Slug:        "data-analyst",
			Disabled:    false,
			Description: "Data analysis specialist with reporting and analytics access",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "Business Analyst",
			Slug:        "business-analyst",
			Disabled:    false,
			Description: "Business analysis with process optimization access",
		},
	},

	// ========== Quality Assurance Roles ==========
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "QA Manager",
			Slug:        "qa-manager",
			Disabled:    false,
			Description: "Quality assurance management with testing oversight",
		},
	},
	{
		RoleBody: accessStructs.RoleBody{
			Name:        "QA Tester",
			Slug:        "qa-tester",
			Disabled:    false,
			Description: "Quality assurance testing with bug reporting access",
		},
	},
}

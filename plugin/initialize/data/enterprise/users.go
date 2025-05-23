package data

import (
	userStructs "ncobase/user/structs"

	"github.com/ncobase/ncore/utils/convert"
)

// SystemDefaultUsers defines enterprise system users with proper role assignments
var SystemDefaultUsers = []UserCreationInfo{
	// ========== System Level Users ==========
	{
		User: userStructs.UserBody{
			Username:    "super",
			Email:       "super@enterprise.com",
			Phone:       "13800138000",
			IsCertified: true,
			IsAdmin:     true,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Super Administrator",
			FirstName:   "Super",
			LastName:    "Admin",
		},
		Role: "super-admin",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP000",
			Position:       "Super Administrator",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-01-01T09:00:00Z"),
		},
	},
	{
		User: userStructs.UserBody{
			Username:    "admin",
			Email:       "admin@enterprise.com",
			Phone:       "13800138010",
			IsCertified: true,
			IsAdmin:     true,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "System Administrator",
			FirstName:   "System",
			LastName:    "Admin",
		},
		Role: "system-admin",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP001",
			Position:       "System Administrator",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-01-01T09:00:00Z"),
		},
	},

	// ========== Enterprise Executives ==========
	{
		User: userStructs.UserBody{
			Username:    "chief.executive",
			Email:       "ceo@enterprise.com",
			Phone:       "13800138001",
			IsCertified: true,
			IsAdmin:     true,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Chief Executive Officer",
			FirstName:   "John",
			LastName:    "Smith",
			Title:       "CEO",
		},
		Role: "enterprise-admin",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP002",
			Department:     "executive",
			Position:       "Chief Executive Officer",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-01-01T09:00:00Z"),
		},
	},

	// ========== Department Managers ==========
	{
		User: userStructs.UserBody{
			Username:    "hr.manager",
			Email:       "hr.manager@enterprise.com",
			Phone:       "13800138002",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "HR Manager",
			FirstName:   "Sarah",
			LastName:    "Johnson",
			Title:       "Human Resources Manager",
		},
		Role: "hr-manager",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP003",
			Department:     "human-resources",
			Position:       "HR Manager",
			ManagerID:      "chief.executive",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-01-15T09:00:00Z"),
		},
	},
	{
		User: userStructs.UserBody{
			Username:    "finance.manager",
			Email:       "finance.manager@enterprise.com",
			Phone:       "13800138003",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Finance Manager",
			FirstName:   "Michael",
			LastName:    "Chen",
			Title:       "Chief Financial Officer",
		},
		Role: "finance-manager",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP004",
			Department:     "finance",
			Position:       "Finance Manager",
			ManagerID:      "chief.executive",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-01-15T09:00:00Z"),
		},
	},
	{
		User: userStructs.UserBody{
			Username:    "it.manager",
			Email:       "it.manager@enterprise.com",
			Phone:       "13800138004",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "IT Manager",
			FirstName:   "David",
			LastName:    "Wilson",
			Title:       "Information Technology Manager",
		},
		Role: "it-manager",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP005",
			Department:     "information-technology",
			Position:       "IT Manager",
			ManagerID:      "chief.executive",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-01-20T09:00:00Z"),
		},
	},

	// ========== Technical Team ==========
	{
		User: userStructs.UserBody{
			Username:    "tech.lead",
			Email:       "tech.lead@techcorp.com",
			Phone:       "13800138005",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Technical Lead",
			FirstName:   "Emily",
			LastName:    "Rodriguez",
			Title:       "Senior Technical Lead",
		},
		Role: "technical-lead",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP006",
			Department:     "technology",
			Position:       "Technical Lead",
			ManagerID:      "it.manager",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-02-01T09:00:00Z"),
			Skills:         &[]string{"System Architecture", "Team Leadership", "Go", "React", "DevOps"},
		},
	},
	{
		User: userStructs.UserBody{
			Username:    "senior.developer",
			Email:       "senior.dev@techcorp.com",
			Phone:       "13800138006",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Senior Developer",
			FirstName:   "Alex",
			LastName:    "Thompson",
			Title:       "Senior Software Engineer",
		},
		Role: "developer",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP007",
			Department:     "technology",
			Position:       "Senior Developer",
			ManagerID:      "tech.lead",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-02-15T09:00:00Z"),
			Skills:         &[]string{"Go", "React", "TypeScript", "PostgreSQL", "Docker", "Kubernetes"},
		},
	},
	{
		User: userStructs.UserBody{
			Username:    "ui.designer",
			Email:       "ui.designer@techcorp.com",
			Phone:       "13800138007",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "UI/UX Designer",
			FirstName:   "Jessica",
			LastName:    "Lee",
			Title:       "Senior UI/UX Designer",
		},
		Role: "designer",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP008",
			Department:     "technology",
			Position:       "UI/UX Designer",
			ManagerID:      "tech.lead",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-02-20T09:00:00Z"),
			Skills:         &[]string{"Figma", "Adobe Creative Suite", "User Research", "Prototyping", "Design Systems"},
		},
	},

	// ========== Marketing Team ==========
	{
		User: userStructs.UserBody{
			Username:    "marketing.manager",
			Email:       "marketing.manager@mediacorp.com",
			Phone:       "13800138008",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Marketing Manager",
			FirstName:   "Lisa",
			LastName:    "Wang",
			Title:       "Digital Marketing Manager",
		},
		Role: "department-manager",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP009",
			Department:     "marketing",
			Position:       "Marketing Manager",
			ManagerID:      "chief.executive",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-02-01T09:00:00Z"),
			Skills:         &[]string{"Digital Marketing", "Content Strategy", "Analytics", "Campaign Management"},
		},
	},
	{
		User: userStructs.UserBody{
			Username:    "content.creator",
			Email:       "content.creator@mediacorp.com",
			Phone:       "13800138009",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Content Creator",
			FirstName:   "Ryan",
			LastName:    "Davis",
			Title:       "Senior Content Creator",
		},
		Role: "content-creator",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP010",
			Department:     "marketing",
			Position:       "Content Creator",
			ManagerID:      "marketing.manager",
			EmploymentType: "contract",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-03-01T09:00:00Z"),
			Skills:         &[]string{"Video Production", "Graphic Design", "Social Media", "Photography", "Adobe Premiere"},
		},
	},

	// ========== Quality Assurance Team ==========
	{
		User: userStructs.UserBody{
			Username:    "qa.manager",
			Email:       "qa.manager@techcorp.com",
			Phone:       "13800138010",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "QA Manager",
			FirstName:   "Jennifer",
			LastName:    "Brown",
			Title:       "Quality Assurance Manager",
		},
		Role: "qa-manager",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP011",
			Department:     "technology",
			Position:       "QA Manager",
			ManagerID:      "tech.lead",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-02-10T09:00:00Z"),
			Skills:         &[]string{"Test Management", "Automation", "Quality Processes", "Team Leadership"},
		},
	},
	{
		User: userStructs.UserBody{
			Username:    "qa.tester",
			Email:       "qa.tester@techcorp.com",
			Phone:       "13800138011",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "QA Tester",
			FirstName:   "Kevin",
			LastName:    "Martinez",
			Title:       "Software QA Engineer",
		},
		Role: "qa-tester",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP012",
			Department:     "technology",
			Position:       "QA Tester",
			ManagerID:      "qa.manager",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-03-15T09:00:00Z"),
			Skills:         &[]string{"Manual Testing", "Selenium", "API Testing", "Bug Reporting", "Test Planning"},
		},
	},

	// ========== Data Analytics Team ==========
	{
		User: userStructs.UserBody{
			Username:    "data.analyst",
			Email:       "data.analyst@enterprise.com",
			Phone:       "13800138012",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Data Analyst",
			FirstName:   "Amanda",
			LastName:    "Garcia",
			Title:       "Senior Data Analyst",
		},
		Role: "data-analyst",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP013",
			Department:     "analytics",
			Position:       "Data Analyst",
			ManagerID:      "finance.manager",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-02-25T09:00:00Z"),
			Skills:         &[]string{"SQL", "Python", "Tableau", "Statistical Analysis", "Data Visualization"},
		},
	},

	// ========== Customer Service Team ==========
	{
		User: userStructs.UserBody{
			Username:    "cs.manager",
			Email:       "cs.manager@enterprise.com",
			Phone:       "13800138013",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Customer Service Manager",
			FirstName:   "Robert",
			LastName:    "Taylor",
			Title:       "Customer Success Manager",
		},
		Role: "customer-service-manager",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP014",
			Department:     "customer-service",
			Position:       "Customer Service Manager",
			ManagerID:      "chief.executive",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-01-30T09:00:00Z"),
		},
	},

	// ========== Regular Employees ==========
	{
		User: userStructs.UserBody{
			Username:    "employee.one",
			Email:       "employee.one@enterprise.com",
			Phone:       "13800138014",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Employee One",
			FirstName:   "Mark",
			LastName:    "Johnson",
			Title:       "Software Engineer",
		},
		Role: "employee",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP015",
			Department:     "technology",
			Position:       "Software Engineer",
			ManagerID:      "tech.lead",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-03-10T09:00:00Z"),
			Skills:         &[]string{"JavaScript", "Node.js", "React", "SQL"},
		},
	},

	// ========== Interns and Contractors ==========
	{
		User: userStructs.UserBody{
			Username:    "intern.dev",
			Email:       "intern.dev@enterprise.com",
			Phone:       "13800138015",
			IsCertified: false,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Development Intern",
			FirstName:   "Sophie",
			LastName:    "Anderson",
			Title:       "Software Development Intern",
		},
		Role: "intern",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "INT001",
			Department:     "technology",
			Position:       "Software Development Intern",
			ManagerID:      "senior.developer",
			EmploymentType: "intern",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-03-20T09:00:00Z"),
			Skills:         &[]string{"Python", "HTML", "CSS", "Git"},
		},
	},
	{
		User: userStructs.UserBody{
			Username:    "contractor.consultant",
			Email:       "contractor@external.com",
			Phone:       "13800138016",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "External Consultant",
			FirstName:   "James",
			LastName:    "Miller",
			Title:       "Business Process Consultant",
		},
		Role: "contractor",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "CON001",
			Department:     "consulting",
			Position:       "Business Consultant",
			EmploymentType: "contract",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-03-05T09:00:00Z"),
			Skills:         &[]string{"Process Optimization", "Business Analysis", "Strategy", "Change Management"},
		},
	},
}

// UserCreationInfo combines user data with related information for initialization
type UserCreationInfo struct {
	User     userStructs.UserBody        `json:"user"`
	Password string                      `json:"password"`
	Profile  userStructs.UserProfileBody `json:"profile"`
	Role     string                      `json:"role"`
	Employee *userStructs.EmployeeBody   `json:"employee,omitempty"`
}

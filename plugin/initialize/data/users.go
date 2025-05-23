package data

import (
	userStructs "ncobase/user/structs"

	"github.com/ncobase/ncore/utils/convert"
)

// SystemDefaultUsers defines enterprise system users
var SystemDefaultUsers = []UserCreationInfo{
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
			Username:    "tech.lead",
			Email:       "tech.lead@techcorp.com",
			Phone:       "13800138004",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Technical Lead",
		},
		Role: "department-manager",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP005",
			Department:     "technology",
			Position:       "Technical Lead",
			ManagerID:      "chief.executive",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-02-01T09:00:00Z"),
		},
	},
	{
		User: userStructs.UserBody{
			Username:    "senior.developer",
			Email:       "senior.dev@techcorp.com",
			Phone:       "13800138005",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Senior Developer",
		},
		Role: "employee",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP006",
			Department:     "technology",
			Position:       "Senior Developer",
			ManagerID:      "tech.lead",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-02-15T09:00:00Z"),
			Skills:         &[]string{"Go", "React", "PostgreSQL", "Docker"},
		},
	},
	{
		User: userStructs.UserBody{
			Username:    "marketing.manager",
			Email:       "marketing.manager@mediacorp.com",
			Phone:       "13800138006",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Marketing Manager",
		},
		Role: "department-manager",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP007",
			Department:     "marketing",
			Position:       "Marketing Manager",
			ManagerID:      "chief.executive",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-02-01T09:00:00Z"),
		},
	},
	{
		User: userStructs.UserBody{
			Username:    "content.creator",
			Email:       "content.creator@mediacorp.com",
			Phone:       "13800138007",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Content Creator",
		},
		Role: "employee",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP008",
			Department:     "marketing",
			Position:       "Content Creator",
			ManagerID:      "marketing.manager",
			EmploymentType: "contract",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-03-01T09:00:00Z"),
			Skills:         &[]string{"Video Production", "Graphic Design", "Social Media"},
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

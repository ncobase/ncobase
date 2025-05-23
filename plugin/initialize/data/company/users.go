package company

import (
	userStructs "ncobase/user/structs"

	"github.com/ncobase/ncore/utils/convert"
)

// SystemDefaultUsers defines core system users
var SystemDefaultUsers = []UserCreationInfo{
	// Super Administrator
	{
		User: userStructs.UserBody{
			Username:    "super",
			Email:       "super@ncobase.com",
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

	// System Administrator
	{
		User: userStructs.UserBody{
			Username:    "admin",
			Email:       "admin@ncobase.com",
			Phone:       "13800138001",
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

	// Enterprise Administrator
	{
		User: userStructs.UserBody{
			Username:    "enterprise.admin",
			Email:       "enterprise.admin@ncobase.com",
			Phone:       "13800138002",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Enterprise Administrator",
			FirstName:   "Enterprise",
			LastName:    "Admin",
		},
		Role: "enterprise-admin",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP002",
			Position:       "Enterprise Administrator",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-01-15T09:00:00Z"),
		},
	},

	// Manager
	{
		User: userStructs.UserBody{
			Username:    "manager",
			Email:       "manager@ncobase.com",
			Phone:       "13800138003",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Manager123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Department Manager",
			FirstName:   "John",
			LastName:    "Manager",
		},
		Role: "manager",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP003",
			Position:       "Department Manager",
			ManagerID:      "enterprise.admin",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-02-01T09:00:00Z"),
		},
	},

	// Employee
	{
		User: userStructs.UserBody{
			Username:    "employee",
			Email:       "employee@ncobase.com",
			Phone:       "13800138004",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Employee123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Employee",
			FirstName:   "Jane",
			LastName:    "Employee",
		},
		Role: "employee",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP004",
			Position:       "Employee",
			ManagerID:      "manager",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-03-01T09:00:00Z"),
		},
	},

	// Guest User
	{
		User: userStructs.UserBody{
			Username:    "guest",
			Email:       "guest@ncobase.com",
			Phone:       "13800138005",
			IsCertified: false,
			IsAdmin:     false,
		},
		Password: "Guest123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Guest User",
			FirstName:   "Guest",
			LastName:    "User",
		},
		Role: "guest",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP005",
			Position:       "Guest",
			EmploymentType: "contract",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-03-15T09:00:00Z"),
		},
	},
}

// UserCreationInfo combines user data for initialization
type UserCreationInfo struct {
	User     userStructs.UserBody        `json:"user"`
	Password string                      `json:"password"`
	Profile  userStructs.UserProfileBody `json:"profile"`
	Role     string                      `json:"role"`
	Employee *userStructs.EmployeeBody   `json:"employee,omitempty"`
}

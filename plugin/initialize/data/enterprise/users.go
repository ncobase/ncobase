package enterprise

import (
	userStructs "ncobase/user/structs"

	"github.com/ncobase/ncore/utils/convert"
)

// SystemDefaultUsers defines simplified enterprise system users
var SystemDefaultUsers = []UserCreationInfo{
	// System Level Users
	{
		User: userStructs.UserBody{
			Username:    "super",
			Email:       "super@enterprise.com",
			Phone:       "13800138000",
			IsCertified: true,
			IsAdmin:     true,
		},
		Password: "Super123456",
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
		Password: "Admin123456",
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

	// Enterprise Management
	{
		User: userStructs.UserBody{
			Username:    "enterprise.admin",
			Email:       "enterprise.admin@enterprise.com",
			Phone:       "13800138001",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Enterprise123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Enterprise Administrator",
			FirstName:   "Enterprise",
			LastName:    "Admin",
		},
		Role: "enterprise-admin",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP002",
			Department:     "executive",
			Position:       "Enterprise Administrator",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-01-01T09:00:00Z"),
		},
	},

	// Department Manager
	{
		User: userStructs.UserBody{
			Username:    "dept.manager",
			Email:       "dept.manager@enterprise.com",
			Phone:       "13800138002",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Manager123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Department Manager",
			FirstName:   "John",
			LastName:    "Manager",
		},
		Role: "department-manager",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP003",
			Department:     "technology",
			Position:       "Department Manager",
			ManagerID:      "enterprise.admin",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-01-15T09:00:00Z"),
		},
	},

	// Team Leader
	{
		User: userStructs.UserBody{
			Username:    "team.leader",
			Email:       "team.leader@enterprise.com",
			Phone:       "13800138003",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Leader123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Team Leader",
			FirstName:   "Sarah",
			LastName:    "Leader",
		},
		Role: "team-leader",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "EMP004",
			Department:     "technology",
			Position:       "Team Leader",
			ManagerID:      "dept.manager",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-02-01T09:00:00Z"),
		},
	},

	// Employee
	{
		User: userStructs.UserBody{
			Username:    "employee",
			Email:       "employee@enterprise.com",
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
			EmployeeID:     "EMP005",
			Department:     "technology",
			Position:       "Software Developer",
			ManagerID:      "team.leader",
			EmploymentType: "full_time",
			Status:         "active",
			HireDate:       convert.ParseTimePtr("2024-03-01T09:00:00Z"),
		},
	},

	// Contractor
	{
		User: userStructs.UserBody{
			Username:    "contractor",
			Email:       "contractor@external.com",
			Phone:       "13800138005",
			IsCertified: false,
			IsAdmin:     false,
		},
		Password: "Contractor123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "External Contractor",
			FirstName:   "Mike",
			LastName:    "Contractor",
		},
		Role: "contractor",
		Employee: &userStructs.EmployeeBody{
			EmployeeID:     "CON001",
			Department:     "consulting",
			Position:       "External Consultant",
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

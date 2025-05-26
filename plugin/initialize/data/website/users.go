package website

import (
	userStructs "ncobase/user/structs"
)

// SystemDefaultUsers for regular websites
var SystemDefaultUsers = []UserCreationInfo{
	// Super Administrator
	{
		User: userStructs.UserBody{
			Username:    "super",
			Email:       "super@website.com",
			Phone:       "13800138000",
			IsCertified: true,
			IsAdmin:     true,
		},
		Password: "Super123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Super Admin",
			FirstName:   "Super",
			LastName:    "Admin",
		},
		Role: "super-admin",
	},

	// Administrator
	{
		User: userStructs.UserBody{
			Username:    "admin",
			Email:       "admin@website.com",
			Phone:       "13800138001",
			IsCertified: true,
			IsAdmin:     true,
		},
		Password: "Admin123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Administrator",
			FirstName:   "Site",
			LastName:    "Admin",
		},
		Role: "admin",
	},

	// Manager
	{
		User: userStructs.UserBody{
			Username:    "manager",
			Email:       "manager@website.com",
			Phone:       "13800138002",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Manager123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Content Manager",
			FirstName:   "Content",
			LastName:    "Manager",
		},
		Role: "manager",
	},

	// Member
	{
		User: userStructs.UserBody{
			Username:    "member",
			Email:       "member@website.com",
			Phone:       "13800138003",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Member123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Member User",
			FirstName:   "Member",
			LastName:    "User",
		},
		Role: "member",
	},

	// Viewer
	{
		User: userStructs.UserBody{
			Username:    "viewer",
			Email:       "viewer@website.com",
			Phone:       "13800138004",
			IsCertified: false,
			IsAdmin:     false,
		},
		Password: "Viewer123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Viewer User",
			FirstName:   "Visitor",
			LastName:    "User",
		},
		Role: "viewer",
	},
}

type UserCreationInfo struct {
	User     userStructs.UserBody        `json:"user"`
	Password string                      `json:"password"`
	Profile  userStructs.UserProfileBody `json:"profile"`
	Role     string                      `json:"role"`
	Employee *userStructs.EmployeeBody   `json:"employee,omitempty"`
}

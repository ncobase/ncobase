package data

import userStructs "ncobase/core/user/structs"

// SystemDefaultUsers defines the default system users
var SystemDefaultUsers = []UserCreationInfo{
	{
		User: userStructs.UserBody{
			Username:    "super",
			Email:       "super@example.com",
			Phone:       "13800000100",
			IsCertified: true,
			IsAdmin:     true,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Super Administrator",
			// Title:       "System Architect",
			// Department:  "IT Department",
		},
		Role: "super-admin",
	},
	{
		User: userStructs.UserBody{
			Username:    "admin",
			Email:       "admin@example.com",
			Phone:       "13800000101",
			IsCertified: true,
			IsAdmin:     true,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "System Administrator",
			// Title:       "IT Manager",
			// Department:  "IT Department",
		},
		Role: "admin",
	},
	{
		User: userStructs.UserBody{
			Username:    "user",
			Email:       "user@example.com",
			Phone:       "13800000102",
			IsCertified: true,
			IsAdmin:     false,
		},
		Password: "Ac123456",
		Profile: userStructs.UserProfileBody{
			DisplayName: "Test User",
			// Title:       "Staff",
			// Department:  "Marketing",
		},
		Role: "user",
	},
}

// UserCreationInfo combines user data with related information for initialization
type UserCreationInfo struct {
	User     userStructs.UserBody        `json:"user"`
	Password string                      `json:"password"`
	Profile  userStructs.UserProfileBody `json:"profile"`
	Role     string                      `json:"role"`
}

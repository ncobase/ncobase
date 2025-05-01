package data

import accessStructs "ncobase/core/access/structs"

// SystemDefaultPermissions defines the system default permissions
var SystemDefaultPermissions = []accessStructs.CreatePermissionBody{
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "All Access",
			Action:      "*",
			Subject:     "*",
			Description: "Full access to all resources",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Read All",
			Action:      "GET",
			Subject:     "*",
			Description: "Read access to all resources",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Create All",
			Action:      "POST",
			Subject:     "*",
			Description: "Create access to all resources",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Update All",
			Action:      "PUT",
			Subject:     "*",
			Description: "Update access to all resources",
		},
	},
	{
		PermissionBody: accessStructs.PermissionBody{
			Name:        "Delete All",
			Action:      "DELETE",
			Subject:     "*",
			Description: "Delete access to all resources",
		},
	},
}

// RolePermissionMapping defines which permissions are assigned to which roles
var RolePermissionMapping = map[string][]string{
	"super-admin": {"All Access"},
	"admin":       {"Read All", "Create All", "Update All", "Delete All"},
	"user":        {"Read All"},
}

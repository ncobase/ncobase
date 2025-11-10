package constants

// AdminRoles contains the list of administrative roles
var AdminRoles = []string{
	"super-admin",
	"system-admin",
	"enterprise-admin",
	"space-admin",
}

// IsAdminRole checks if a role is an admin role
func IsAdminRole(role string) bool {
	for _, adminRole := range AdminRoles {
		if role == adminRole {
			return true
		}
	}
	return false
}

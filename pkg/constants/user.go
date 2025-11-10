package constants

// UserStatus represents the status of a user
type UserStatus int

const (
	UserStatusActive UserStatus = iota
	UserStatusInactive
	UserStatusPending
)

// String returns the string representation of UserStatus
func (s UserStatus) String() string {
	switch s {
	case UserStatusActive:
		return "active"
	case UserStatusInactive:
		return "inactive"
	case UserStatusPending:
		return "pending"
	default:
		return "unknown"
	}
}

// ParseUserStatus parses a string into UserStatus
func ParseUserStatus(status string) UserStatus {
	switch status {
	case "active":
		return UserStatusActive
	case "inactive":
		return UserStatusInactive
	case "pending":
		return UserStatusPending
	default:
		return UserStatusActive
	}
}

// ToInt converts UserStatus to int
func (s UserStatus) ToInt() int {
	return int(s)
}

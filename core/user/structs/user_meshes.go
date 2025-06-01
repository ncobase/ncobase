package structs

// UserMeshes represents aggregated user information
type UserMeshes struct {
	User     *ReadUser        `json:"user,omitempty"`
	Profile  *ReadUserProfile `json:"profile,omitempty"`
	Employee *ReadEmployee    `json:"employee,omitempty"`
	ApiKeys  []*ApiKey        `json:"api_keys,omitempty"`
}

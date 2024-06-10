package structs

// CasbinRuleParams defines the structure for query parameters used to find Casbin rules.
type CasbinRuleParams struct {
	PType  *string `form:"p_type"`
	V0     *string `form:"v0"`
	V1     *string `form:"v1"`
	V2     *string `form:"v2"`
	V3     *string `form:"v3"`
	V4     *string `form:"v4"`
	V5     *string `form:"v5"`
	Limit  int     `form:"limit"`
	Offset int     `form:"offset"`
}

// CasbinRuleBody defines the structure for request body used to create or update Casbin rules.
type CasbinRuleBody struct {
	PType string `json:"p_type" binding:"required"`
	V0    string `json:"v0" binding:"required"`
	V1    string `json:"v1" binding:"required"`
	V2    string `json:"v2" binding:"required"`
	V3    string `json:"v3"`
	V4    string `json:"v4"`
	V5    string `json:"v5"`
}

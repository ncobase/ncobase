package service

import (
	accessStructs "ncobase/core/access/structs"
	companyData "ncobase/plugin/initialize/data/company"
	enterpriseData "ncobase/plugin/initialize/data/enterprise"
	websiteData "ncobase/plugin/initialize/data/website"
	spaceStructs "ncobase/core/space/structs"
	systemStructs "ncobase/core/system/structs"
	userStructs "ncobase/core/user/structs"
)

// DataLoader interface for mode-specific data
type DataLoader interface {
	GetRoles() []accessStructs.CreateRoleBody
	GetPermissions() []accessStructs.CreatePermissionBody
	GetRolePermissionMapping() map[string][]string
	GetCasbinPolicyRules() [][]string
	GetRoleInheritanceRules() [][]string
	GetUsers() []UserCreationInfo
	GetSpaces() []spaceStructs.CreateSpaceBody
	GetSpaceQuotas() []spaceStructs.CreateSpaceQuotaBody
	GetSpaceSettings() []spaceStructs.CreateSpaceSettingBody
	GetOptions() []systemStructs.OptionBody
	GetDictionaries() []systemStructs.DictionaryBody
	GetOrganizationStructure() any
}

// UserCreationInfo combines user data for initialization
type UserCreationInfo struct {
	User     userStructs.UserBody        `json:"user"`
	Password string                      `json:"password"`
	Profile  userStructs.UserProfileBody `json:"profile"`
	Role     string                      `json:"role"`
	Employee *userStructs.EmployeeBody   `json:"employee,omitempty"`
}

// WebsiteDataLoader for website mode
type WebsiteDataLoader struct{}

func (w *WebsiteDataLoader) GetRoles() []accessStructs.CreateRoleBody {
	return websiteData.SystemDefaultRoles
}

func (w *WebsiteDataLoader) GetPermissions() []accessStructs.CreatePermissionBody {
	return websiteData.SystemDefaultPermissions
}

func (w *WebsiteDataLoader) GetRolePermissionMapping() map[string][]string {
	return websiteData.RolePermissionMapping
}

func (w *WebsiteDataLoader) GetCasbinPolicyRules() [][]string {
	return websiteData.CasbinPolicyRules
}

func (w *WebsiteDataLoader) GetRoleInheritanceRules() [][]string {
	return websiteData.RoleInheritanceRules
}

func (w *WebsiteDataLoader) GetUsers() []UserCreationInfo {
	users := make([]UserCreationInfo, len(websiteData.SystemDefaultUsers))
	for i, u := range websiteData.SystemDefaultUsers {
		users[i] = UserCreationInfo{
			User:     u.User,
			Password: u.Password,
			Profile:  u.Profile,
			Role:     u.Role,
			Employee: u.Employee,
		}
	}
	return users
}

func (w *WebsiteDataLoader) GetSpaces() []spaceStructs.CreateSpaceBody {
	return websiteData.SystemDefaultSpaces
}

func (w *WebsiteDataLoader) GetSpaceQuotas() []spaceStructs.CreateSpaceQuotaBody {
	return websiteData.SystemDefaultSpaceQuotas
}

func (w *WebsiteDataLoader) GetSpaceSettings() []spaceStructs.CreateSpaceSettingBody {
	return websiteData.SystemDefaultSpaceSettings
}

func (w *WebsiteDataLoader) GetOptions() []systemStructs.OptionBody {
	return websiteData.SystemDefaultOptions
}

func (w *WebsiteDataLoader) GetDictionaries() []systemStructs.DictionaryBody {
	return websiteData.SystemDefaultDictionaries
}

func (w *WebsiteDataLoader) GetOrganizationStructure() any {
	return websiteData.OrganizationStructure
}

// CompanyDataLoader for company mode
type CompanyDataLoader struct{}

func (c *CompanyDataLoader) GetRoles() []accessStructs.CreateRoleBody {
	return companyData.SystemDefaultRoles
}

func (c *CompanyDataLoader) GetPermissions() []accessStructs.CreatePermissionBody {
	return companyData.SystemDefaultPermissions
}

func (c *CompanyDataLoader) GetRolePermissionMapping() map[string][]string {
	return companyData.RolePermissionMapping
}

func (c *CompanyDataLoader) GetCasbinPolicyRules() [][]string {
	return companyData.CasbinPolicyRules
}

func (c *CompanyDataLoader) GetRoleInheritanceRules() [][]string {
	return companyData.RoleInheritanceRules
}

func (c *CompanyDataLoader) GetUsers() []UserCreationInfo {
	users := make([]UserCreationInfo, len(companyData.SystemDefaultUsers))
	for i, u := range companyData.SystemDefaultUsers {
		users[i] = UserCreationInfo{
			User:     u.User,
			Password: u.Password,
			Profile:  u.Profile,
			Role:     u.Role,
			Employee: u.Employee,
		}
	}
	return users
}

func (c *CompanyDataLoader) GetSpaces() []spaceStructs.CreateSpaceBody {
	return companyData.SystemDefaultSpaces
}

func (c *CompanyDataLoader) GetSpaceQuotas() []spaceStructs.CreateSpaceQuotaBody {
	return companyData.SystemDefaultSpaceQuotas
}

func (c *CompanyDataLoader) GetSpaceSettings() []spaceStructs.CreateSpaceSettingBody {
	return companyData.SystemDefaultSpaceSettings
}

func (c *CompanyDataLoader) GetOptions() []systemStructs.OptionBody {
	return companyData.SystemDefaultOptions
}

func (c *CompanyDataLoader) GetDictionaries() []systemStructs.DictionaryBody {
	return companyData.SystemDefaultDictionaries
}

func (c *CompanyDataLoader) GetOrganizationStructure() any {
	return companyData.OrganizationStructure
}

// EnterpriseDataLoader for enterprise mode
type EnterpriseDataLoader struct{}

func (e *EnterpriseDataLoader) GetRoles() []accessStructs.CreateRoleBody {
	return enterpriseData.SystemDefaultRoles
}

func (e *EnterpriseDataLoader) GetPermissions() []accessStructs.CreatePermissionBody {
	return enterpriseData.SystemDefaultPermissions
}

func (e *EnterpriseDataLoader) GetRolePermissionMapping() map[string][]string {
	return enterpriseData.RolePermissionMapping
}

func (e *EnterpriseDataLoader) GetCasbinPolicyRules() [][]string {
	return enterpriseData.CasbinPolicyRules
}

func (e *EnterpriseDataLoader) GetRoleInheritanceRules() [][]string {
	return enterpriseData.RoleInheritanceRules
}

func (e *EnterpriseDataLoader) GetUsers() []UserCreationInfo {
	users := make([]UserCreationInfo, len(enterpriseData.SystemDefaultUsers))
	for i, u := range enterpriseData.SystemDefaultUsers {
		users[i] = UserCreationInfo{
			User:     u.User,
			Password: u.Password,
			Profile:  u.Profile,
			Role:     u.Role,
			Employee: u.Employee,
		}
	}
	return users
}

func (e *EnterpriseDataLoader) GetSpaces() []spaceStructs.CreateSpaceBody {
	return enterpriseData.SystemDefaultSpaces
}

func (e *EnterpriseDataLoader) GetSpaceQuotas() []spaceStructs.CreateSpaceQuotaBody {
	return enterpriseData.SystemDefaultSpaceQuotas
}

func (e *EnterpriseDataLoader) GetSpaceSettings() []spaceStructs.CreateSpaceSettingBody {
	return enterpriseData.SystemDefaultSpaceSettings
}

func (e *EnterpriseDataLoader) GetOptions() []systemStructs.OptionBody {
	return enterpriseData.SystemDefaultOptions
}

func (e *EnterpriseDataLoader) GetDictionaries() []systemStructs.DictionaryBody {
	return enterpriseData.SystemDefaultDictionaries
}

func (e *EnterpriseDataLoader) GetOrganizationStructure() any {
	return enterpriseData.OrganizationStructure
}

// getDataLoader returns appropriate data loader based on current mode
func (s *Service) getDataLoader() DataLoader {
	switch s.state.DataMode {
	case "enterprise":
		return &EnterpriseDataLoader{}
	case "company":
		return &CompanyDataLoader{}
	case "website":
		return &WebsiteDataLoader{}
	default:
		return &WebsiteDataLoader{}
	}
}

package service

import (
	accessStructs "ncobase/access/structs"
	menuData "ncobase/initialize/data"
	companyData "ncobase/initialize/data/company"
	enterpriseData "ncobase/initialize/data/enterprise"
	websiteData "ncobase/initialize/data/website"
	systemStructs "ncobase/system/structs"
	tenantStructs "ncobase/tenant/structs"
	userStructs "ncobase/user/structs"
)

type DataLoader interface {
	GetRoles() []accessStructs.CreateRoleBody
	GetPermissions() []accessStructs.CreatePermissionBody
	GetRolePermissionMapping() map[string][]string
	GetCasbinPolicyRules() [][]string
	GetRoleInheritanceRules() [][]string
	GetUsers() []UserCreationInfo
	GetTenants() []tenantStructs.CreateTenantBody
	GetOptions() []systemStructs.OptionsBody
	GetDictionaries() []systemStructs.DictionaryBody
	GetOrganizationStructure() any
}

type UserCreationInfo struct {
	User     userStructs.UserBody        `json:"user"`
	Password string                      `json:"password"`
	Profile  userStructs.UserProfileBody `json:"profile"`
	Role     string                      `json:"role"`
	Employee *userStructs.EmployeeBody   `json:"employee,omitempty"`
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

func (c *CompanyDataLoader) GetTenants() []tenantStructs.CreateTenantBody {
	return companyData.SystemDefaultTenants
}

func (c *CompanyDataLoader) GetOptions() []systemStructs.OptionsBody {
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

func (e *EnterpriseDataLoader) GetTenants() []tenantStructs.CreateTenantBody {
	return enterpriseData.SystemDefaultTenants
}

func (e *EnterpriseDataLoader) GetOptions() []systemStructs.OptionsBody {
	return enterpriseData.SystemDefaultOptions
}

func (e *EnterpriseDataLoader) GetDictionaries() []systemStructs.DictionaryBody {
	return enterpriseData.SystemDefaultDictionaries
}

func (e *EnterpriseDataLoader) GetOrganizationStructure() any {
	return enterpriseData.OrganizationStructure
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

func (w *WebsiteDataLoader) GetTenants() []tenantStructs.CreateTenantBody {
	return websiteData.SystemDefaultTenants
}

func (w *WebsiteDataLoader) GetOptions() []systemStructs.OptionsBody {
	return websiteData.SystemDefaultOptions
}

func (w *WebsiteDataLoader) GetDictionaries() []systemStructs.DictionaryBody {
	return websiteData.SystemDefaultDictionaries
}

func (w *WebsiteDataLoader) GetOrganizationStructure() any {
	return websiteData.OrganizationStructure
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

func (s *Service) getMenuData() *struct {
	Headers  []systemStructs.MenuBody
	Sidebars []systemStructs.MenuBody
	Submenus []systemStructs.MenuBody
	Accounts []systemStructs.MenuBody
	Tenants  []systemStructs.MenuBody
} {
	return &menuData.SystemDefaultMenus
}

package handler

import "ncobase/tenant/service"

// Handler represents the tenant handler
type Handler struct {
	Tenant           TenantHandlerInterface
	TenantQuota      TenantQuotaHandlerInterface
	TenantSetting    TenantSettingHandlerInterface
	TenantBilling    TenantBillingHandlerInterface
	UserTenantRole   UserTenantRoleHandlerInterface
	TenantGroup      TenantGroupHandlerInterface
	TenantMenu       TenantMenuHandlerInterface
	TenantDictionary TenantDictionaryHandlerInterface
	TenantOptions    TenantOptionsHandlerInterface
}

// New creates a new handler
func New(svc *service.Service) *Handler {
	return &Handler{
		Tenant:           NewTenantHandler(svc),
		TenantQuota:      NewTenantQuotaHandler(svc),
		TenantSetting:    NewTenantSettingHandler(svc),
		TenantBilling:    NewTenantBillingHandler(svc),
		UserTenantRole:   NewUserTenantRoleHandler(svc),
		TenantGroup:      NewTenantGroupHandler(svc),
		TenantMenu:       NewTenantMenuHandler(svc),
		TenantDictionary: NewTenantDictionaryHandler(svc),
		TenantOptions:    NewTenantOptionsHandler(svc),
	}
}

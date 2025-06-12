package handler

import "ncobase/space/service"

// Handler represents the space handler
type Handler struct {
	Space             SpaceHandlerInterface
	SpaceQuota        SpaceQuotaHandlerInterface
	SpaceSetting      SpaceSettingHandlerInterface
	SpaceBilling      SpaceBillingHandlerInterface
	UserSpaceRole     UserSpaceRoleHandlerInterface
	SpaceOrganization SpaceOrganizationHandlerInterface
	SpaceMenu         SpaceMenuHandlerInterface
	SpaceDictionary   SpaceDictionaryHandlerInterface
	SpaceOption       SpaceOptionHandlerInterface
}

// New creates a new handler
func New(svc *service.Service) *Handler {
	return &Handler{
		Space:             NewSpaceHandler(svc),
		SpaceQuota:        NewSpaceQuotaHandler(svc),
		SpaceSetting:      NewSpaceSettingHandler(svc),
		SpaceBilling:      NewSpaceBillingHandler(svc),
		UserSpaceRole:     NewUserSpaceRoleHandler(svc),
		SpaceOrganization: NewSpaceOrganizationHandler(svc),
		SpaceMenu:         NewSpaceMenuHandler(svc),
		SpaceDictionary:   NewSpaceDictionaryHandler(svc),
		SpaceOption:       NewSpaceOptionHandler(svc),
	}
}

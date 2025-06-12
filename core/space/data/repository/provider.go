package repository

import (
	"ncobase/space/data"
)

// Repository represents the space repository
type Repository struct {
	Space             SpaceRepositoryInterface
	UserSpace         UserSpaceRepositoryInterface
	UserSpaceRole     UserSpaceRoleRepositoryInterface
	SpaceQuota        SpaceQuotaRepositoryInterface
	SpaceSetting      SpaceSettingRepositoryInterface
	SpaceBilling      SpaceBillingRepositoryInterface
	SpaceOrganization SpaceOrganizationRepositoryInterface
	SpaceMenu         SpaceMenuRepositoryInterface
	SpaceDictionary   SpaceDictionaryRepositoryInterface
	SpaceOption       SpaceOptionRepositoryInterface
}

// New creates a new repository
func New(d *data.Data) *Repository {
	return &Repository{
		Space:             NewSpaceRepository(d),
		UserSpace:         NewUserSpaceRepository(d),
		UserSpaceRole:     NewUserSpaceRoleRepository(d),
		SpaceQuota:        NewSpaceQuotaRepository(d),
		SpaceSetting:      NewSpaceSettingRepository(d),
		SpaceBilling:      NewSpaceBillingRepository(d),
		SpaceOrganization: NewSpaceOrganizationRepository(d),
		SpaceMenu:         NewSpaceMenuRepository(d),
		SpaceDictionary:   NewSpaceDictionaryRepository(d),
		SpaceOption:       NewSpaceOptionRepository(d),
	}
}

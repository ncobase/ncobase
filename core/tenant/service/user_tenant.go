package service

import (
	"context"
	"errors"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	"ncobase/tenant/data/repository"
	"ncobase/tenant/structs"

	"github.com/ncobase/ncore/ecode"
)

// UserTenantServiceInterface is the interface for the service.
type UserTenantServiceInterface interface {
	UserBelongTenant(ctx context.Context, uid string) (*structs.ReadTenant, error)
	UserBelongTenants(ctx context.Context, uid string) ([]*structs.ReadTenant, error)
	AddUserToTenant(ctx context.Context, u, t string) (*structs.UserTenant, error)
	RemoveUserFromTenant(ctx context.Context, u, t string) error
	IsTenantInUser(ctx context.Context, t, u string) (bool, error)
}

// userTenantService is the struct for the service.
type userTenantService struct {
	ts         TenantServiceInterface
	userTenant repository.UserTenantRepositoryInterface
}

// NewUserTenantService creates a new service.
func NewUserTenantService(d *data.Data, ts TenantServiceInterface) UserTenantServiceInterface {
	return &userTenantService{
		ts:         ts,
		userTenant: repository.NewUserTenantRepository(d),
	}
}

// UserBelongTenant user belong tenant service
func (s *userTenantService) UserBelongTenant(ctx context.Context, uid string) (*structs.ReadTenant, error) {
	if uid == "" {
		return nil, errors.New(ecode.FieldIsInvalid("User ID"))
	}

	// Try to get tenant from user-tenant relationship
	userTenant, err := s.userTenant.GetByUserID(ctx, uid)
	if err != nil {
		// If no specific tenant found, try to get the first available tenant for the user
		tenants, err := s.userTenant.GetTenantsByUserID(ctx, uid)
		if err != nil || len(tenants) == 0 {
			// If user doesn't belong to any tenant, check if they created a tenant
			return s.ts.GetByUser(ctx, uid)
		}
		// Return the first tenant
		return s.ts.Serialize(tenants[0]), nil
	}

	row, err := s.ts.Find(ctx, userTenant.TenantID)
	if err := handleEntError(ctx, "Tenant", err); err != nil {
		return nil, err
	}

	return row, nil
}

// UserBelongTenants user belong tenants service
func (s *userTenantService) UserBelongTenants(ctx context.Context, uid string) ([]*structs.ReadTenant, error) {
	if uid == "" {
		return nil, errors.New(ecode.FieldIsInvalid("User ID"))
	}

	userTenants, err := s.userTenant.GetTenantsByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}

	var tenants []*structs.ReadTenant
	for _, userTenant := range userTenants {
		tenant, err := s.ts.Find(ctx, userTenant.ID)
		if err != nil {
			return nil, errors.New("tenant not found")
		}
		tenants = append(tenants, tenant)
	}

	return tenants, nil
}

// AddUserToTenant adds a user to a tenant.
func (s *userTenantService) AddUserToTenant(ctx context.Context, u string, t string) (*structs.UserTenant, error) {
	row, err := s.userTenant.Create(ctx, &structs.UserTenant{UserID: u, TenantID: t})
	if err := handleEntError(ctx, "UserTenant", err); err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// RemoveUserFromTenant removes a user from a tenant.
func (s *userTenantService) RemoveUserFromTenant(ctx context.Context, u string, t string) error {
	err := s.userTenant.Delete(ctx, u, t)
	if err := handleEntError(ctx, "UserTenant", err); err != nil {
		return err
	}
	return nil
}

// IsTenantInUser checks if a tenant is in a user.
func (s *userTenantService) IsTenantInUser(ctx context.Context, t, u string) (bool, error) {
	isValid, err := s.userTenant.IsTenantInUser(ctx, t, u)
	if err = handleEntError(ctx, "UserTenant", err); err != nil {
		return false, err

	}
	return isValid, nil
}

// Serializes serializes user tenants.
func (s *userTenantService) Serializes(rows []*ent.UserTenant) []*structs.UserTenant {
	rs := make([]*structs.UserTenant, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes a user tenant.
func (s *userTenantService) Serialize(row *ent.UserTenant) *structs.UserTenant {
	return &structs.UserTenant{
		UserID:   row.UserID,
		TenantID: row.TenantID,
	}
}

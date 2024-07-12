package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
	"ncobase/feature/tenant/data"
	"ncobase/feature/tenant/data/ent"
	"ncobase/feature/tenant/data/repository"
	"ncobase/feature/tenant/structs"
)

// UserTenantServiceInterface is the interface for the service.
type UserTenantServiceInterface interface {
	UserBelongTenant(ctx context.Context, uid string) (*structs.ReadTenant, error)
	UserBelongTenants(ctx context.Context, uid string) ([]*structs.ReadTenant, error)
	AddUserToTenant(ctx context.Context, u string, t string) (*structs.UserTenant, error)
	RemoveUserFromTenant(ctx context.Context, u string, t string) error
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

	userTenant, err := s.userTenant.GetByUserID(ctx, uid)
	if err := handleEntError("UserTenant", err); err != nil {
		return nil, err
	}

	row, err := s.ts.Find(ctx, userTenant.TenantID)
	if err := handleEntError("Tenant", err); err != nil {
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
	if err := handleEntError("UserTenant", err); err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// RemoveUserFromTenant removes a user from a tenant.
func (s *userTenantService) RemoveUserFromTenant(ctx context.Context, u string, t string) error {
	err := s.userTenant.Delete(ctx, u, t)
	if err := handleEntError("UserTenant", err); err != nil {
		return err
	}
	return nil
}

// Serializes serializes user tenants.
func (s *userTenantService) Serializes(rows []*ent.UserTenant) []*structs.UserTenant {
	var rs []*structs.UserTenant
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

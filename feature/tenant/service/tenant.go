package service

import (
	"context"
	"encoding/json"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/helper"
	"ncobase/common/log"
	"ncobase/common/paging"
	"ncobase/common/types"
	"ncobase/feature/tenant/data"
	"ncobase/feature/tenant/data/ent"
	"ncobase/feature/tenant/data/repository"
	"ncobase/feature/tenant/structs"
)

// TenantServiceInterface is the interface for the service.
type TenantServiceInterface interface {
	UserOwn(ctx context.Context, uid string) (*structs.ReadTenant, error)
	Create(ctx context.Context, body *structs.CreateTenantBody) (*structs.ReadTenant, error)
	Update(ctx context.Context, body *structs.UpdateTenantBody) (*structs.ReadTenant, error)
	Get(ctx context.Context, id string) (*structs.ReadTenant, error)
	GetBySlug(ctx context.Context, id string) (*structs.ReadTenant, error)
	GetByUser(ctx context.Context, uid string) (*structs.ReadTenant, error)
	Find(ctx context.Context, id string) (*structs.ReadTenant, error)
	Delete(ctx context.Context, id string) error
	CountX(ctx context.Context, params *structs.ListTenantParams) int
	List(ctx context.Context, params *structs.ListTenantParams) (paging.Result[*structs.ReadTenant], error)
	Serializes(rows []*ent.Tenant) []*structs.ReadTenant
	Serialize(tenant *ent.Tenant) *structs.ReadTenant
}

// tenantService is the struct for the service.
type tenantService struct {
	tenant     repository.TenantRepositoryInterface
	userTenant repository.UserTenantRepositoryInterface
}

// NewTenantService creates a new service.
func NewTenantService(d *data.Data) TenantServiceInterface {
	return &tenantService{
		tenant:     repository.NewTenantRepository(d),
		userTenant: repository.NewUserTenantRepository(d),
	}
}

// UserOwn user own tenant service
func (s *tenantService) UserOwn(ctx context.Context, uid string) (*structs.ReadTenant, error) {
	if uid == "" {
		return nil, errors.New(ecode.FieldIsInvalid("User ID"))
	}

	row, err := s.tenant.GetByUser(ctx, uid)
	if err := handleEntError(ctx, "Tenant", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Create creates a tenant service.
func (s *tenantService) Create(ctx context.Context, body *structs.CreateTenantBody) (*structs.ReadTenant, error) {
	if body.CreatedBy == nil {
		body.CreatedBy = types.ToPointer(helper.GetUserID(ctx))
	}

	// Create the tenant
	tenant, err := s.tenant.Create(ctx, body)
	if err != nil {
		return nil, err
	}

	return s.Serialize(tenant), nil
}

// Update updates tenant service (full and partial).
func (s *tenantService) Update(ctx context.Context, body *structs.UpdateTenantBody) (*structs.ReadTenant, error) {
	userID := helper.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// Check if CreatedBy field is provided and validate user's access to the tenant
	if body.CreatedBy != nil {
		_, err := s.tenant.GetByUser(ctx, *body.CreatedBy)
		if err := handleEntError(ctx, "Tenant", err); err != nil {
			return nil, err
		}
	}

	// If ID is not provided, get the tenant ID associated with the user
	if body.ID == "" {
		body.ID, _ = s.tenant.GetIDByUser(ctx, userID)
	}

	// Retrieve the tenant by ID
	row, err := s.Find(ctx, body.ID)
	if err := handleEntError(ctx, "Tenant", err); err != nil {
		return nil, err
	}

	// Check if the user is the creator or user belongs to the tenant
	ut, err := s.userTenant.GetByUserID(ctx, userID)
	if err := handleEntError(ctx, "UserTenant", err); err != nil {
		return nil, err
	}
	if types.ToValue(row.CreatedBy) != userID && ut.TenantID != row.ID {
		return nil, errors.New("this tenant is not yours or your not belong to this tenant")
	}

	// set updated by
	body.UpdatedBy = &userID

	// Serialize request body
	bodyData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data into a types.JSON object
	var d types.JSON
	if err := json.Unmarshal(bodyData, &d); err != nil {
		return nil, err
	}

	// Update the tenant with the provided data
	_, err = s.tenant.Update(ctx, row.ID, d)
	if err := handleEntError(ctx, "Tenant", err); err != nil {
		return nil, err
	}

	return row, nil
}

// Get reads tenant service.
func (s *tenantService) Get(ctx context.Context, id string) (*structs.ReadTenant, error) {
	userID := helper.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// If ID is not provided, get the tenant ID associated with the user
	if id == "" {
		id, _ = s.tenant.GetIDByUser(ctx, userID)
	}

	// Retrieve the tenant by ID
	row, err := s.Find(ctx, id)
	if err := handleEntError(ctx, "Tenant", err); err != nil {
		return nil, err
	}

	// Check if the user is the creator or user belongs to the tenant
	ut, err := s.userTenant.GetByUserID(ctx, userID)
	if err := handleEntError(ctx, "UserTenant", err); err != nil {
		return nil, err
	}
	if types.ToValue(row.CreatedBy) != userID && ut.TenantID != row.ID {
		return nil, errors.New("this tenant is not yours or your not belong to this tenant")
	}

	return row, nil
}

// GetBySlug returns the tenant for the provided slug
func (s *tenantService) GetBySlug(ctx context.Context, slug string) (*structs.ReadTenant, error) {
	if slug == "" {
		return nil, errors.New(ecode.FieldIsInvalid("Slug"))
	}
	tenant, err := s.tenant.GetBySlug(ctx, slug)
	if err := handleEntError(ctx, "Tenant", err); err != nil {
		return nil, err
	}
	return s.Serialize(tenant), nil
}

// GetByUser returns the tenant for the created by user
func (s *tenantService) GetByUser(ctx context.Context, uid string) (*structs.ReadTenant, error) {
	if uid == "" {
		return nil, errors.New(ecode.FieldIsInvalid("User ID"))
	}
	tenant, err := s.tenant.GetByUser(ctx, uid)
	if err := handleEntError(ctx, "Tenant", err); err != nil {
		return nil, err
	}
	return s.Serialize(tenant), nil
}

// Find finds tenant service.
func (s *tenantService) Find(ctx context.Context, id string) (*structs.ReadTenant, error) {
	tenant, err := s.tenant.GetBySlug(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.Serialize(tenant), nil
}

// Delete deletes tenant service.
func (s *tenantService) Delete(ctx context.Context, id string) error {
	err := s.tenant.Delete(ctx, id)
	if err != nil {
		return err
	}

	// TODO: Freed all roles / groups / users that are associated with the tenant

	return nil
}

// List lists tenant service.
func (s *tenantService) List(ctx context.Context, params *structs.ListTenantParams) (paging.Result[*structs.ReadTenant], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadTenant, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.tenant.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			log.Errorf(ctx, "Error listing tenants: %v\n", err)
			return nil, 0, err
		}

		total := s.CountX(ctx, params)

		return s.Serializes(rows), total, nil
	})
}

// CountX gets a count of tenants.
func (s *tenantService) CountX(ctx context.Context, params *structs.ListTenantParams) int {
	return s.tenant.CountX(ctx, params)
}

// Serializes serialize tenants
func (s *tenantService) Serializes(rows []*ent.Tenant) []*structs.ReadTenant {
	tenants := make([]*structs.ReadTenant, 0, len(rows))
	for _, row := range rows {
		tenants = append(tenants, s.Serialize(row))
	}
	return tenants
}

// Serialize serialize a tenant
func (s *tenantService) Serialize(row *ent.Tenant) *structs.ReadTenant {
	return &structs.ReadTenant{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Type:        row.Type,
		Title:       row.Title,
		URL:         row.URL,
		Logo:        row.Logo,
		LogoAlt:     row.LogoAlt,
		Keywords:    row.Keywords,
		Copyright:   row.Copyright,
		Description: row.Description,
		Order:       &row.Order,
		Disabled:    row.Disabled,
		Extras:      &row.Extras,
		ExpiredAt:   &row.ExpiredAt,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}
}

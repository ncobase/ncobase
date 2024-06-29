package service

import (
	"context"
	"encoding/json"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/slug"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/internal/data/ent"
	"ncobase/internal/data/structs"
	"ncobase/internal/helper"

	"github.com/gin-gonic/gin"
)

// AccountTenantService retrieves the tenant associated with the user's account.
func (svc *Service) AccountTenantService(c *gin.Context) (*resp.Exception, error) {
	userID := helper.GetUserID(c)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// Retrieve the tenant associated with the user
	tenant, err := svc.tenant.GetByUser(c, userID)
	if exception, err := helper.HandleError("Tenant", err); exception != nil {
		return exception, err
	}

	// Serialize tenant data and return
	return &resp.Exception{
		Data: svc.serializeTenant(tenant),
	}, nil
}

// AccountTenantsService retrieves the tenant associated with the user's account.
func (svc *Service) AccountTenantsService(c *gin.Context) (*resp.Exception, error) {
	userID := helper.GetUserID(c)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	tenants, err := svc.ListTenantsService(c, &structs.ListTenantParams{
		User: userID,
	})
	if exception, err := helper.HandleError("Tenants", err); exception != nil {
		return exception, err
	}

	return tenants, nil
}

// UserOwnTenantService user own tenant service
func (svc *Service) UserOwnTenantService(c *gin.Context, username string) (*resp.Exception, error) {
	if username == "" {
		return resp.BadRequest(ecode.FieldIsInvalid("username")), nil
	}

	user, err := svc.findUser(c, &structs.FindUser{Username: username})
	if exception, err := helper.HandleError("User", err); exception != nil {
		return exception, err
	}

	tenant, err := svc.tenant.GetByUser(c, user.User.ID)
	if exception, err := helper.HandleError("Tenant", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeTenant(tenant),
	}, nil
}

// UserBelongTenantService user belong tenant service
func (svc *Service) UserBelongTenantService(c *gin.Context, username string) (*resp.Exception, error) {
	if username == "" {
		return resp.BadRequest(ecode.FieldIsInvalid("username")), nil
	}

	user, err := svc.findUser(c, &structs.FindUser{Username: username})
	if exception, err := helper.HandleError("User", err); exception != nil {
		return exception, err
	}

	userTenant, err := svc.userTenant.GetByUserID(c, user.User.ID)
	if exception, err := helper.HandleError("UserTenant", err); exception != nil {
		return exception, err
	}

	tenant, err := svc.tenant.GetBySlug(c, userTenant.TenantID)
	if exception, err := helper.HandleError("Tenant", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeTenant(tenant),
	}, nil
}

// UserBelongTenantsService user belong tenants service
func (svc *Service) UserBelongTenantsService(c *gin.Context, username string) (*resp.Exception, error) {
	if username == "" {
		return resp.BadRequest(ecode.FieldIsInvalid("username")), nil
	}

	user, err := svc.findUser(c, &structs.FindUser{Username: username})
	if exception, err := helper.HandleError("User", err); exception != nil {
		return exception, err
	}

	userTenants, err := svc.userTenant.GetTenantsByUserID(c, user.User.ID)
	if exception, err := helper.HandleError("UserTenants", err); exception != nil {
		return exception, err
	}

	var tenants []*ent.Tenant
	for _, userTenant := range userTenants {
		tenant, err := svc.tenant.GetBySlug(c, userTenant.ID)
		if exception, err := helper.HandleError("Tenant", err); exception != nil {
			return exception, err
		}
		tenants = append(tenants, tenant)
	}

	return &resp.Exception{
		Data: svc.serializeTenants(tenants),
	}, nil
}

// CreateTenantService creates a tenant service.
func (svc *Service) CreateTenantService(c *gin.Context, body *structs.CreateTenantBody) (*resp.Exception, error) {
	if body.CreatedBy == nil {
		body.CreatedBy = types.ToPointer(helper.GetUserID(c))
	}

	// Create the tenant
	tenant, err := svc.isCreateTenant(helper.FromGinContext(c), body)
	if exception, err := helper.HandleError("Tenant", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeTenant(tenant),
	}, nil
}

// UpdateTenantService updates tenant service (full and partial).
func (svc *Service) UpdateTenantService(c *gin.Context, body *structs.UpdateTenantBody) (*resp.Exception, error) {
	userID := helper.GetUserID(c)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// Check if CreatedBy field is provided and validate user's access to the tenant
	if body.CreatedBy != nil {
		_, err := svc.tenant.GetByUser(helper.FromGinContext(c), *body.CreatedBy)
		if exception, err := helper.HandleError("Tenant", err); exception != nil {
			return exception, err
		}
	}

	// If ID is not provided, get the tenant ID associated with the user
	if body.ID == "" {
		body.ID, _ = svc.tenant.GetIDByUser(helper.FromGinContext(c), userID)
	}

	// Retrieve the tenant by ID
	tenant, err := svc.tenant.GetBySlug(helper.FromGinContext(c), body.ID)
	if exception, err := helper.HandleError("Tenant", err); exception != nil {
		return exception, err
	}

	// Check if the user is the creator of the tenant
	if tenant.CreatedBy != userID {
		return resp.Forbidden("This tenant is not yours"), nil
	}

	// set updated by
	body.UpdatedBy = &userID

	// Serialize request body
	bodyData, err := json.Marshal(body)
	if err != nil {
		return resp.InternalServer(err.Error()), nil
	}

	// Unmarshal the JSON data into a types.JSON object
	var data types.JSON
	if err := json.Unmarshal(bodyData, &data); err != nil {
		return resp.InternalServer(err.Error()), nil
	}

	// Update the tenant with the provided data
	_, err = svc.tenant.Update(helper.FromGinContext(c), tenant.ID, data)
	if exception, err := helper.HandleError("Tenant", err); exception != nil {
		return exception, err
	}

	// Return success response
	return &resp.Exception{
		Data: types.JSON{
			"id": body.ID,
		},
	}, nil
}

// GetTenantService reads tenant service.
func (svc *Service) GetTenantService(c *gin.Context, id string) (*resp.Exception, error) {
	userID := helper.GetUserID(c)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// If ID is not provided, get the tenant ID associated with the user
	if id == "" {
		id, _ = svc.tenant.GetIDByUser(helper.FromGinContext(c), userID)
	}

	// Retrieve the tenant by ID
	tenant, err := svc.tenant.GetBySlug(helper.FromGinContext(c), id)
	if exception, err := helper.HandleError("Tenant", err); exception != nil {
		return exception, err
	}

	// Check if the user is the creator of the tenant
	if tenant.CreatedBy != userID {
		return resp.Forbidden("This tenant is not yours"), nil
	}

	// Serialize tenant data and return
	return &resp.Exception{
		Data: svc.serializeTenant(tenant),
	}, nil
}

// DeleteTenantService deletes tenant service.
func (svc *Service) DeleteTenantService(c *gin.Context, id string) (*resp.Exception, error) {
	err := svc.tenant.Delete(c, id)
	if err != nil {
		return resp.BadRequest(err.Error()), nil
	}

	// TODO: Freed all roles / groups / users that are associated with the tenant

	return &resp.Exception{
		Message: "Tenant deleted successfully",
	}, nil
}

// ListTenantsService lists tenant service.
func (svc *Service) ListTenantsService(c *gin.Context, params *structs.ListTenantParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	tenants, err := svc.tenant.List(helper.FromGinContext(c), params)
	if exception, err := helper.HandleError("Tenant", err); exception != nil {
		return exception, err
	}

	total := svc.tenant.CountX(helper.FromGinContext(c), params)

	return &resp.Exception{
		Data: types.JSON{
			"content": tenants,
			"total":   total,
		},
	}, nil
}

// ****** Internal methods of service

// createInitialTenant creates the initial tenant, initializes roles, and user relationships.
func (svc *Service) createInitialTenant(ctx context.Context, body *structs.CreateTenantBody) (*ent.Tenant, error) {
	// Create the default tenant
	defaultTenant, err := svc.tenant.Create(ctx, body)
	if err != nil {
		return nil, err
	}

	// Get or create the super admin role
	superAdminRole, err := svc.role.GetBySlug(ctx, "super-admin")
	if ent.IsNotFound(err) {
		// Super admin role does not exist, create it
		superAdminRole, err = svc.role.CreateSuperAdminRole(ctx)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Assign the user to the default tenant with the super admin role
	if body.CreatedBy != nil {
		_, err = svc.userTenant.Create(ctx, &structs.UserTenant{UserID: *body.CreatedBy, TenantID: defaultTenant.ID})
		if err != nil {
			return nil, err
		}

		// Assign the tenant to the super admin role
		_, err = svc.userTenantRole.Create(ctx, &structs.UserTenantRole{UserID: *body.CreatedBy, RoleID: superAdminRole.ID, TenantID: defaultTenant.ID})
		if err != nil {
			return nil, err
		}

		// Assign the super admin role to the user
		_, err = svc.userRole.Create(ctx, &structs.UserRole{UserID: *body.CreatedBy, RoleID: superAdminRole.ID})
		if err != nil {
			return nil, err
		}
	}

	return defaultTenant, nil
}

// isCreateTenant checks if a tenant needs to be created and initializes tenant, roles, and user relationships if necessary.
func (svc *Service) isCreateTenant(ctx context.Context, body *structs.CreateTenantBody) (*ent.Tenant, error) {
	if body.CreatedBy == nil {
		return nil, errors.New("invalid user ID")
	}

	// If slug is not provided, generate it
	if body.Slug == "" && body.Name != "" {
		body.Slug = slug.Unicode(body.Name)
	}

	// Check the number of existing users
	countUsers := svc.user.CountX(ctx, &structs.ListUserParams{})

	// If there are no existing users, create the initial tenant
	if countUsers <= 1 {
		return svc.createInitialTenant(ctx, body)
	}

	// If there are existing users, check if the user already has a tenant
	existingTenant, err := svc.tenant.GetByUser(ctx, *body.CreatedBy)
	if ent.IsNotFound(err) {
		// No existing tenant found for the user, proceed with tenant creation
	} else if err != nil {
		return nil, err
	} else {
		// If the user already has a tenant, return the existing tenant
		return existingTenant, nil
	}

	// If there are no existing tenants and body.Tenant is not empty, create the initial tenant
	if body.TenantBody.Name != "" {
		return svc.createInitialTenant(ctx, body)
	}

	return nil, nil

}

// serializeTenants serialize tenants
func (svc *Service) serializeTenants(rows []*ent.Tenant) []*structs.ReadTenant {
	tenants := make([]*structs.ReadTenant, 0, len(rows))
	for _, row := range rows {
		tenants = append(tenants, svc.serializeTenant(row))
	}
	return tenants
}

// serializeTenant serialize tenant
func (svc *Service) serializeTenant(row *ent.Tenant) *structs.ReadTenant {
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
		BaseEntity: structs.BaseEntity{
			CreatedBy: &row.CreatedBy,
			CreatedAt: &row.CreatedAt,
			UpdatedBy: &row.UpdatedBy,
			UpdatedAt: &row.UpdatedAt,
		},
	}
}

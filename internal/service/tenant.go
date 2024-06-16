package service

import (
	"context"
	"encoding/json"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/internal/data/ent"
	"ncobase/internal/data/structs"
	"ncobase/internal/helper"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

// AccountTenantService retrieves the tenant associated with the user's account.
func (svc *Service) AccountTenantService(c *gin.Context) (*resp.Exception, error) {
	userID := helper.GetUserID(c)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// Retrieve the tenant associated with the user
	tenant, err := svc.tenant.GetByUser(c, userID)
	if exception, err := handleError("Tenant", err); exception != nil {
		return exception, err
	}

	// Serialize tenant data and return
	return &resp.Exception{
		Data: svc.serializeTenant(c, tenant, false),
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
	if exception, err := handleError("Tenants", err); exception != nil {
		return exception, err
	}

	return tenants, nil
}

// UserTenantService user tenant service
func (svc *Service) UserTenantService(c *gin.Context, username string) (*resp.Exception, error) {
	if username == "" {
		return resp.BadRequest(ecode.FieldIsInvalid("username")), nil
	}

	user, err := svc.findUser(c, &structs.FindUser{Username: username})
	if exception, err := handleError("User", err); exception != nil {
		return exception, err
	}

	tenant, err := svc.tenant.GetByUser(c, user.ID)
	if exception, err := handleError("Tenant", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeTenant(c, tenant, false),
	}, nil
}

// CreateTenantService creates a tenant service.
func (svc *Service) CreateTenantService(c *gin.Context, body *structs.CreateTenantBody) (*resp.Exception, error) {
	if body.CreatedBy == "" {
		body.CreatedBy = helper.GetUserID(c)
	}

	// Create the tenant
	tenant, err := svc.isCreateTenant(helper.FromGinContext(c), body)
	if exception, err := handleError("Tenant", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeTenant(c, tenant, true),
	}, nil
}

// UpdateTenantService updates tenant service (full and partial).
func (svc *Service) UpdateTenantService(c *gin.Context, body *structs.UpdateTenantBody) (*resp.Exception, error) {
	userID := helper.GetUserID(c)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// Check if CreatedBy field is provided and validate user's access to the tenant
	if body.CreatedBy != "" {
		_, err := svc.tenant.GetByUser(helper.FromGinContext(c), body.CreatedBy)
		if exception, err := handleError("Tenant", err); exception != nil {
			return exception, err
		}
	}

	// If ID is not provided, get the tenant ID associated with the user
	if body.ID == "" {
		body.ID, _ = svc.tenant.GetIDByUser(helper.FromGinContext(c), userID)
	}

	// Retrieve the tenant by ID
	tenant, err := svc.tenant.GetByID(helper.FromGinContext(c), body.ID)
	if exception, err := handleError("Tenant", err); exception != nil {
		return exception, err
	}

	// Check if the user is the creator of the tenant
	if tenant.CreatedBy != userID {
		return resp.Forbidden("This tenant is not yours"), nil
	}

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
	if exception, err := handleError("Tenant", err); exception != nil {
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
	tenant, err := svc.tenant.GetByID(helper.FromGinContext(c), id)
	if exception, err := handleError("Tenant", err); exception != nil {
		return exception, err
	}

	// Check if the user is the creator of the tenant
	if tenant.CreatedBy != userID {
		return resp.Forbidden("This tenant is not yours"), nil
	}

	// Serialize tenant data and return
	return &resp.Exception{
		Data: svc.serializeTenant(c, tenant, true),
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
	if exception, err := handleError("Tenant", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: tenants,
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
	_, err = svc.userTenant.Create(ctx, &structs.UserTenant{UserID: body.CreatedBy, TenantID: defaultTenant.ID})
	if err != nil {
		return nil, err
	}

	// Assign the tenant to the super admin role
	_, err = svc.userTenantRole.Create(ctx, &structs.UserTenantRole{UserID: body.CreatedBy, RoleID: superAdminRole.ID, TenantID: defaultTenant.ID})
	if err != nil {
		return nil, err
	}

	// Assign the super admin role to the user
	_, err = svc.userRole.Create(ctx, &structs.UserRole{UserID: body.CreatedBy, RoleID: superAdminRole.ID})
	if err != nil {
		return nil, err
	}

	return defaultTenant, nil
}

// isCreateTenant checks if a tenant needs to be created and initializes tenant, roles, and user relationships if necessary.
func (svc *Service) isCreateTenant(ctx context.Context, body *structs.CreateTenantBody) (*ent.Tenant, error) {
	if body.CreatedBy == "" {
		return nil, errors.New("invalid user ID")
	}

	client := svc.d.GetEntClient()

	// Check the number of existing users
	countUsers, err := client.User.Query().Count(ctx)
	if err != nil {
		return nil, err
	}

	// If there are no existing users, create the initial tenant
	if countUsers <= 1 {
		return svc.createInitialTenant(ctx, body)
	}

	// If there are existing users, check if the user already has a tenant
	existingTenant, err := svc.tenant.GetByUser(ctx, body.CreatedBy)
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

// serializeTenant serialize tenant
func (svc *Service) serializeTenant(c *gin.Context, row *ent.Tenant, withUser bool) *structs.ReadTenant {
	readTenant := &structs.ReadTenant{
		ID:          row.ID,
		Name:        row.Name,
		Title:       row.Title,
		URL:         row.URL,
		Logo:        row.Logo,
		LogoAlt:     row.LogoAlt,
		Keywords:    strings.Split(row.Keywords, ","),
		Copyright:   row.Copyright,
		Description: row.Description,
		Order:       &row.Order,
		Disabled:    row.Disabled,
		Extras:      &row.Extras,
		CreatedBy:   row.CreatedBy,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}

	if withUser {
		user, err := svc.user.GetByID(c, row.CreatedBy)
		if err == nil {
			readTenant.User = new(structs.User)
			_ = copier.Copy(&readTenant.User, user)
		}
	}

	return readTenant
}

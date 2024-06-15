package service

import (
	"context"
	"encoding/json"
	"errors"
	"ncobase/internal/data/ent"
	"ncobase/internal/data/structs"
	"ncobase/internal/helper"
	"ncobase/pkg/ecode"
	"ncobase/pkg/resp"
	"ncobase/pkg/types"
	"ncobase/pkg/validator"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

// AccountDomainService retrieves the domain associated with the user's account.
func (svc *Service) AccountDomainService(c *gin.Context) (*resp.Exception, error) {
	userID := helper.GetUserID(c)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// Retrieve the domain associated with the user
	domain, err := svc.domain.GetByUser(c, userID)
	if exception, err := handleError("Domain", err); exception != nil {
		return exception, err
	}

	// Serialize domain data and return
	return &resp.Exception{
		Data: svc.serializeDomain(c, domain, false),
	}, nil
}

// UserDomainService user domain service
func (svc *Service) UserDomainService(c *gin.Context, username string) (*resp.Exception, error) {
	if username == "" {
		return resp.BadRequest(ecode.FieldIsInvalid("username")), nil
	}

	user, err := svc.findUser(c, &structs.FindUser{Username: username})
	if exception, err := handleError("User", err); exception != nil {
		return exception, err
	}

	domain, err := svc.domain.GetByUser(c, user.ID)
	if exception, err := handleError("Domain", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeDomain(c, domain, false),
	}, nil
}

// CreateDomainService creates a domain service.
func (svc *Service) CreateDomainService(c *gin.Context, body *structs.CreateDomainBody) (*resp.Exception, error) {
	if body.CreatedBy == "" {
		body.CreatedBy = helper.GetUserID(c)
	}

	// Create the domain
	domain, err := svc.isCreateDomain(helper.FromGinContext(c), body)
	if exception, err := handleError("Domain", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeDomain(c, domain, true),
	}, nil
}

// UpdateDomainService updates domain service (full and partial).
func (svc *Service) UpdateDomainService(c *gin.Context, body *structs.UpdateDomainBody) (*resp.Exception, error) {
	userID := helper.GetUserID(c)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// Check if CreatedBy field is provided and validate user's access to the domain
	if body.CreatedBy != "" {
		_, err := svc.domain.GetByUser(helper.FromGinContext(c), body.CreatedBy)
		if exception, err := handleError("Domain", err); exception != nil {
			return exception, err
		}
	}

	// If ID is not provided, get the domain ID associated with the user
	if body.ID == "" {
		body.ID, _ = svc.domain.GetIDByUser(helper.FromGinContext(c), userID)
	}

	// Retrieve the domain by ID
	domain, err := svc.domain.GetByID(helper.FromGinContext(c), body.ID)
	if exception, err := handleError("Domain", err); exception != nil {
		return exception, err
	}

	// Check if the user is the creator of the domain
	if domain.CreatedBy != userID {
		return resp.Forbidden("This domain is not yours"), nil
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

	// Update the domain with the provided data
	_, err = svc.domain.Update(helper.FromGinContext(c), domain.ID, data)
	if exception, err := handleError("Domain", err); exception != nil {
		return exception, err
	}

	// Return success response
	return &resp.Exception{
		Data: types.JSON{
			"id": body.ID,
		},
	}, nil
}

// GetDomainService reads domain service.
func (svc *Service) GetDomainService(c *gin.Context, id string) (*resp.Exception, error) {
	userID := helper.GetUserID(c)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// If ID is not provided, get the domain ID associated with the user
	if id == "" {
		id, _ = svc.domain.GetIDByUser(helper.FromGinContext(c), userID)
	}

	// Retrieve the domain by ID
	domain, err := svc.domain.GetByID(helper.FromGinContext(c), id)
	if exception, err := handleError("Domain", err); exception != nil {
		return exception, err
	}

	// Check if the user is the creator of the domain
	if domain.CreatedBy != userID {
		return resp.Forbidden("This domain is not yours"), nil
	}

	// Serialize domain data and return
	return &resp.Exception{
		Data: svc.serializeDomain(c, domain, true),
	}, nil
}

// DeleteDomainService deletes domain service.
func (svc *Service) DeleteDomainService(c *gin.Context, id string) (*resp.Exception, error) {
	err := svc.domain.Delete(c, id)
	if err != nil {
		return resp.BadRequest(err.Error()), nil
	}

	// TODO: Freed all roles / groups / users that are associated with the domain

	return &resp.Exception{
		Message: "Domain deleted successfully",
	}, nil
}

// ListDomainsService lists domain service.
func (svc *Service) ListDomainsService(c *gin.Context, params *structs.ListDomainParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	domains, err := svc.domain.List(helper.FromGinContext(c), params)
	if exception, err := handleError("Domain", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: domains,
	}, nil
}

// ****** Internal methods of service

// createInitialDomain creates the initial domain, initializes roles, and user relationships.
func (svc *Service) createInitialDomain(ctx context.Context, body *structs.CreateDomainBody) (*ent.Domain, error) {
	// Create the default domain
	defaultDomain, err := svc.domain.Create(ctx, body)
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

	// Assign the user to the default domain with the super admin role
	_, err = svc.userDomain.Create(ctx, &structs.UserDomain{UserID: body.CreatedBy, DomainID: defaultDomain.ID})
	if err != nil {
		return nil, err
	}

	// Assign the domain to the super admin role
	_, err = svc.userDomainRole.Create(ctx, &structs.UserDomainRole{UserID: body.CreatedBy, RoleID: superAdminRole.ID, DomainID: defaultDomain.ID})
	if err != nil {
		return nil, err
	}

	// Assign the super admin role to the user
	_, err = svc.userRole.Create(ctx, &structs.UserRole{UserID: body.CreatedBy, RoleID: superAdminRole.ID})
	if err != nil {
		return nil, err
	}

	return defaultDomain, nil
}

// isCreateDomain checks if a domain needs to be created and initializes domain, roles, and user relationships if necessary.
func (svc *Service) isCreateDomain(ctx context.Context, body *structs.CreateDomainBody) (*ent.Domain, error) {
	if body.CreatedBy == "" {
		return nil, errors.New("invalid user ID")
	}

	client := svc.d.GetEntClient()

	// Check the number of existing users
	countUsers, err := client.User.Query().Count(ctx)
	if err != nil {
		return nil, err
	}

	// If there are no existing users, create the initial domain
	if countUsers <= 1 {
		return svc.createInitialDomain(ctx, body)
	}

	// If there are existing users, check if the user already has a domain
	existingDomain, err := svc.domain.GetByUser(ctx, body.CreatedBy)
	if ent.IsNotFound(err) {
		// No existing domain found for the user, proceed with domain creation
	} else if err != nil {
		return nil, err
	} else {
		// If the user already has a domain, return the existing domain
		return existingDomain, nil
	}

	// If there are no existing domains and body.Domain is not empty, create the initial domain
	if body.DomainBody.Name != "" {
		return svc.createInitialDomain(ctx, body)
	}

	return nil, nil

}

// serializeDomain serialize domain
func (svc *Service) serializeDomain(c *gin.Context, row *ent.Domain, withUser bool) *structs.ReadDomain {
	readDomain := &structs.ReadDomain{
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
			readDomain.User = new(structs.User)
			_ = copier.Copy(&readDomain.User, user)
		}
	}

	return readDomain
}

package service

import (
	"context"
	"ncobase/initialize/data"
	"ncobase/system/structs"

	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

// checkMenusInitialized Check if menus are initialized
func (s *Service) checkMenusInitialized(ctx context.Context) error {
	params := &structs.ListMenuParams{}
	count := s.sys.Menu.CountX(ctx, params)
	if count > 0 {
		logger.Infof(ctx, "Menus already exist, skipping menu initialization")
		return nil
	}

	return s.initMenus(ctx)
}

// Initialize default menu structure
func (s *Service) initMenus(ctx context.Context) error {
	logger.Infof(ctx, "Initializing default menus...")

	// Get default tenant and admin
	tenant, err := s.ts.Tenant.GetBySlug(ctx, "ncobase")
	if err != nil {
		logger.Errorf(ctx, "initMenus error on get default tenant: %v", err)
		return err
	}

	admin, err := s.us.User.Get(ctx, "admin")
	if err != nil {
		logger.Errorf(ctx, "initMenus error on get admin user: %v", err)
		return err
	}

	defaultExtras := make(types.JSON)

	// Create header menus and map IDs
	headerIDMap := make(map[string]string)
	for _, header := range data.SystemDefaultMenus.Headers {
		menuBody := header // Copy to avoid modifying original
		menuBody.TenantID = tenant.ID
		menuBody.CreatedBy = &admin.ID
		menuBody.UpdatedBy = &admin.ID
		menuBody.Extras = &defaultExtras

		logger.Debugf(ctx, "Creating header menu: %s", menuBody.Name)
		createdMenu, err := s.sys.Menu.Create(ctx, &menuBody)
		if err != nil {
			logger.Errorf(ctx, "Error creating header menu %s: %v", menuBody.Name, err)
			return err
		}
		headerIDMap[menuBody.Slug] = createdMenu.ID
	}

	// Create sidebar menus and map IDs
	sidebarIDMap := make(map[string]string)
	for _, sidebar := range data.SystemDefaultMenus.Sidebars {
		menuBody := sidebar // Copy to avoid modifying original

		// Map parent ID from header slug
		parentID := menuBody.ParentID
		if id, ok := headerIDMap[parentID]; ok {
			menuBody.ParentID = id
		}

		menuBody.TenantID = tenant.ID
		menuBody.CreatedBy = &admin.ID
		menuBody.UpdatedBy = &admin.ID
		menuBody.Extras = &defaultExtras

		logger.Debugf(ctx, "Creating sidebar menu: %s", menuBody.Name)
		createdMenu, err := s.sys.Menu.Create(ctx, &menuBody)
		if err != nil {
			logger.Errorf(ctx, "Error creating sidebar menu %s: %v", menuBody.Name, err)
			return err
		}

		if menuBody.Slug != "" {
			sidebarIDMap[menuBody.Slug] = createdMenu.ID
		}
	}

	// Create submenus
	for _, submenu := range data.SystemDefaultMenus.Submenus {
		menuBody := submenu // Copy to avoid modifying original

		// Map parent ID from sidebar slug
		parentID := menuBody.ParentID
		if id, ok := sidebarIDMap[parentID]; ok {
			menuBody.ParentID = id
		}

		menuBody.TenantID = tenant.ID
		menuBody.CreatedBy = &admin.ID
		menuBody.UpdatedBy = &admin.ID
		menuBody.Extras = &defaultExtras

		logger.Debugf(ctx, "Creating submenu: %s", menuBody.Name)
		if _, err := s.sys.Menu.Create(ctx, &menuBody); err != nil {
			logger.Errorf(ctx, "Error creating submenu %s: %v", menuBody.Name, err)
			return err
		}
	}

	// Create account menus
	for _, menu := range data.SystemDefaultMenus.Accounts {
		menuBody := menu // Copy to avoid modifying original
		menuBody.TenantID = tenant.ID
		menuBody.CreatedBy = &admin.ID
		menuBody.UpdatedBy = &admin.ID
		menuBody.Extras = &defaultExtras

		logger.Debugf(ctx, "Creating account menu: %s", menuBody.Name)
		if _, err := s.sys.Menu.Create(ctx, &menuBody); err != nil {
			logger.Errorf(ctx, "Error creating account menu %s: %v", menuBody.Name, err)
			return err
		}
	}

	// Create tenant menus
	for _, menu := range data.SystemDefaultMenus.Tenants {
		menuBody := menu // Copy to avoid modifying original
		menuBody.TenantID = tenant.ID
		menuBody.CreatedBy = &admin.ID
		menuBody.UpdatedBy = &admin.ID
		menuBody.Extras = &defaultExtras

		logger.Debugf(ctx, "Creating tenant menu: %s", menuBody.Name)
		if _, err := s.sys.Menu.Create(ctx, &menuBody); err != nil {
			logger.Errorf(ctx, "Error creating tenant menu %s: %v", menuBody.Name, err)
			return err
		}
	}

	logger.Infof(ctx, "Menu initialization completed successfully")
	return nil
}

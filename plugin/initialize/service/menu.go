package service

import (
	"context"
	"fmt"
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

// verifyMenuData validates menu data before initialization
func (s *Service) verifyMenuData(ctx context.Context) error {
	// Check headers
	if len(data.SystemDefaultMenus.Headers) == 0 {
		return fmt.Errorf("no header menus defined")
	}

	// Check for required header slugs
	requiredHeaders := map[string]bool{
		"dashboard": false,
		"system":    false,
	}

	for _, header := range data.SystemDefaultMenus.Headers {
		if _, ok := requiredHeaders[header.Slug]; ok {
			requiredHeaders[header.Slug] = true
		}
	}

	for slug, found := range requiredHeaders {
		if !found {
			return fmt.Errorf("required header menu '%s' not defined", slug)
		}
	}

	return nil
}

// Initialize default menu structure
func (s *Service) initMenus(ctx context.Context) error {
	logger.Infof(ctx, "Initializing default menus...")

	// Verify menu data integrity
	if err := s.verifyMenuData(ctx); err != nil {
		return fmt.Errorf("menu data verification failed: %w", err)
	}

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

	// Transaction count for monitoring
	var createdMenus int

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
			return fmt.Errorf("failed to create header menu '%s': %w", menuBody.Name, err)
		}
		headerIDMap[menuBody.Slug] = createdMenu.ID
		createdMenus++
	}

	// Validate that all header menus were created
	for slug := range headerIDMap {
		logger.Debugf(ctx, "Created header menu with slug: %s", slug)
	}

	// Create sidebar menus and map IDs
	sidebarIDMap := make(map[string]string)
	for _, sidebar := range data.SystemDefaultMenus.Sidebars {
		menuBody := sidebar // Copy to avoid modifying original

		// Map parent ID from header slug
		if menuBody.ParentID != "" {
			if id, ok := headerIDMap[menuBody.ParentID]; ok {
				menuBody.ParentID = id
			} else {
				logger.Warnf(ctx, "Parent header '%s' not found for sidebar '%s', skipping",
					menuBody.ParentID, menuBody.Name)
				continue
			}
		}

		menuBody.TenantID = tenant.ID
		menuBody.CreatedBy = &admin.ID
		menuBody.UpdatedBy = &admin.ID
		menuBody.Extras = &defaultExtras

		logger.Debugf(ctx, "Creating sidebar menu: %s", menuBody.Name)
		createdMenu, err := s.sys.Menu.Create(ctx, &menuBody)
		if err != nil {
			logger.Errorf(ctx, "Error creating sidebar menu %s: %v", menuBody.Name, err)
			return fmt.Errorf("failed to create sidebar menu '%s': %w", menuBody.Name, err)
		}

		if menuBody.Slug != "" {
			sidebarIDMap[menuBody.Slug] = createdMenu.ID
		}
		createdMenus++
	}

	// Create submenus
	for _, submenu := range data.SystemDefaultMenus.Submenus {
		menuBody := submenu // Copy to avoid modifying original

		// Map parent ID from sidebar slug
		if menuBody.ParentID != "" {
			if id, ok := sidebarIDMap[menuBody.ParentID]; ok {
				menuBody.ParentID = id
			} else {
				logger.Warnf(ctx, "Parent sidebar '%s' not found for submenu '%s', skipping",
					menuBody.ParentID, menuBody.Name)
				continue
			}
		}

		menuBody.TenantID = tenant.ID
		menuBody.CreatedBy = &admin.ID
		menuBody.UpdatedBy = &admin.ID
		menuBody.Extras = &defaultExtras

		logger.Debugf(ctx, "Creating submenu: %s", menuBody.Name)
		if _, err := s.sys.Menu.Create(ctx, &menuBody); err != nil {
			logger.Errorf(ctx, "Error creating submenu %s: %v", menuBody.Name, err)
			return fmt.Errorf("failed to create submenu '%s': %w", menuBody.Name, err)
		}
		createdMenus++
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
			return fmt.Errorf("failed to create account menu '%s': %w", menuBody.Name, err)
		}
		createdMenus++
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
			return fmt.Errorf("failed to create tenant menu '%s': %w", menuBody.Name, err)
		}
		createdMenus++
	}

	// Verify menu count
	finalCount := s.sys.Menu.CountX(ctx, &structs.ListMenuParams{})
	if finalCount != createdMenus {
		logger.Warnf(ctx, "Menu count mismatch. Expected %d, got %d", createdMenus, finalCount)
	}

	logger.Infof(ctx, "Menu initialization completed successfully. Created %d menus", createdMenus)
	return nil
}

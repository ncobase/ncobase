package service

import (
	"context"
	"fmt"
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
	menuData := s.getMenuData()

	if len(menuData.Headers) == 0 {
		return fmt.Errorf("no header menus defined")
	}

	requiredHeaders := map[string]bool{
		"dashboard": false,
		"system":    false,
	}

	for _, header := range menuData.Headers {
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

// Initialize default menu structure and create tenant relationships
func (s *Service) initMenus(ctx context.Context) error {
	logger.Infof(ctx, "Initializing default menus...")

	if err := s.verifyMenuData(ctx); err != nil {
		return fmt.Errorf("menu data verification failed: %w", err)
	}

	// Get default tenant based on data mode
	var tenantSlug string
	switch s.state.DataMode {
	case "website":
		tenantSlug = "website-platform"
	case "company":
		tenantSlug = "digital-company"
	case "enterprise":
		tenantSlug = "digital-enterprise"
	default:
		tenantSlug = "website-platform"
	}

	tenant, err := s.ts.Tenant.GetBySlug(ctx, tenantSlug)
	if err != nil {
		logger.Errorf(ctx, "initMenus error on get default tenant: %v", err)
		return fmt.Errorf("failed to get default tenant: %w", err)
	}

	adminUser, err := s.getAdminUser(ctx, "menu creation")
	if err != nil {
		return err
	}

	if adminUser == nil {
		logger.Errorf(ctx, "initMenus error: no admin user found")
		return fmt.Errorf("no suitable admin user found for menu creation")
	}

	defaultExtras := make(types.JSON)
	menuData := s.getMenuData()

	var createdMenus, relationshipCount int

	// Create header menus and map IDs
	headerIDMap := make(map[string]string)
	for _, header := range menuData.Headers {
		menuBody := header
		// Remove tenant_id from menu creation
		menuBody.CreatedBy = &adminUser.ID
		menuBody.UpdatedBy = &adminUser.ID
		menuBody.Extras = &defaultExtras

		logger.Debugf(ctx, "Creating header menu: %s", menuBody.Name)
		createdMenu, err := s.sys.Menu.Create(ctx, &menuBody)
		if err != nil {
			logger.Errorf(ctx, "Error creating header menu %s: %v", menuBody.Name, err)
			return fmt.Errorf("failed to create header menu '%s': %w", menuBody.Name, err)
		}
		headerIDMap[menuBody.Slug] = createdMenu.ID
		createdMenus++

		// Create tenant-menu relationship
		_, err = s.ts.TenantMenu.AddMenuToTenant(ctx, tenant.ID, createdMenu.ID)
		if err != nil {
			logger.Errorf(ctx, "Error linking menu %s to tenant %s: %v", createdMenu.ID, tenant.ID, err)
			return err
		}
		relationshipCount++
	}

	// Create sidebar menus and map IDs
	sidebarIDMap := make(map[string]string)
	for _, sidebar := range menuData.Sidebars {
		menuBody := sidebar

		if menuBody.ParentID != "" {
			if id, ok := headerIDMap[menuBody.ParentID]; ok {
				menuBody.ParentID = id
			} else {
				logger.Warnf(ctx, "Parent header '%s' not found for sidebar '%s', skipping",
					menuBody.ParentID, menuBody.Name)
				continue
			}
		}

		// Remove tenant_id from menu creation
		menuBody.CreatedBy = &adminUser.ID
		menuBody.UpdatedBy = &adminUser.ID
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

		// Create tenant-menu relationship
		_, err = s.ts.TenantMenu.AddMenuToTenant(ctx, tenant.ID, createdMenu.ID)
		if err != nil {
			logger.Errorf(ctx, "Error linking menu %s to tenant %s: %v", createdMenu.ID, tenant.ID, err)
			return err
		}
		relationshipCount++
	}

	// Create submenus
	for _, submenu := range menuData.Submenus {
		menuBody := submenu

		if menuBody.ParentID != "" {
			if id, ok := sidebarIDMap[menuBody.ParentID]; ok {
				menuBody.ParentID = id
			} else {
				logger.Warnf(ctx, "Parent sidebar '%s' not found for submenu '%s', skipping",
					menuBody.ParentID, menuBody.Name)
				continue
			}
		}

		// Remove tenant_id from menu creation
		menuBody.CreatedBy = &adminUser.ID
		menuBody.UpdatedBy = &adminUser.ID
		menuBody.Extras = &defaultExtras

		logger.Debugf(ctx, "Creating submenu: %s", menuBody.Name)
		createdMenu, err := s.sys.Menu.Create(ctx, &menuBody)
		if err != nil {
			logger.Errorf(ctx, "Error creating submenu %s: %v", menuBody.Name, err)
			return fmt.Errorf("failed to create submenu '%s': %w", menuBody.Name, err)
		}
		createdMenus++

		// Create tenant-menu relationship
		_, err = s.ts.TenantMenu.AddMenuToTenant(ctx, tenant.ID, createdMenu.ID)
		if err != nil {
			logger.Errorf(ctx, "Error linking menu %s to tenant %s: %v", createdMenu.ID, tenant.ID, err)
			return err
		}
		relationshipCount++
	}

	// Create account menus
	for _, menu := range menuData.Accounts {
		menuBody := menu
		// Remove tenant_id from menu creation
		menuBody.CreatedBy = &adminUser.ID
		menuBody.UpdatedBy = &adminUser.ID
		menuBody.Extras = &defaultExtras

		logger.Debugf(ctx, "Creating account menu: %s", menuBody.Name)
		createdMenu, err := s.sys.Menu.Create(ctx, &menuBody)
		if err != nil {
			logger.Errorf(ctx, "Error creating account menu %s: %v", menuBody.Name, err)
			return fmt.Errorf("failed to create account menu '%s': %w", menuBody.Name, err)
		}
		createdMenus++

		// Create tenant-menu relationship
		_, err = s.ts.TenantMenu.AddMenuToTenant(ctx, tenant.ID, createdMenu.ID)
		if err != nil {
			logger.Errorf(ctx, "Error linking menu %s to tenant %s: %v", createdMenu.ID, tenant.ID, err)
			return err
		}
		relationshipCount++
	}

	// Create tenant menus
	for _, menu := range menuData.Tenants {
		menuBody := menu
		// Remove tenant_id from menu creation
		menuBody.CreatedBy = &adminUser.ID
		menuBody.UpdatedBy = &adminUser.ID
		menuBody.Extras = &defaultExtras

		logger.Debugf(ctx, "Creating tenant menu: %s", menuBody.Name)
		createdMenu, err := s.sys.Menu.Create(ctx, &menuBody)
		if err != nil {
			logger.Errorf(ctx, "Error creating tenant menu %s: %v", menuBody.Name, err)
			return fmt.Errorf("failed to create tenant menu '%s': %w", menuBody.Name, err)
		}
		createdMenus++

		// Create tenant-menu relationship
		_, err = s.ts.TenantMenu.AddMenuToTenant(ctx, tenant.ID, createdMenu.ID)
		if err != nil {
			logger.Errorf(ctx, "Error linking menu %s to tenant %s: %v", createdMenu.ID, tenant.ID, err)
			return err
		}
		relationshipCount++
	}

	finalCount := s.sys.Menu.CountX(ctx, &structs.ListMenuParams{})
	if finalCount != createdMenus {
		logger.Warnf(ctx, "Menu count mismatch. Expected %d, got %d", createdMenus, finalCount)
	}

	logger.Infof(ctx, "Menu initialization completed successfully. Created %d menus and %d relationships using admin user '%s'",
		createdMenus, relationshipCount, adminUser.Username)
	return nil
}

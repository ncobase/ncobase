package service

import (
	"context"
	"fmt"
	"ncobase/core/system/structs"
	menuData "ncobase/plugin/initialize/data"

	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

// checkMenusInitialized Check if menus are initialized
func (s *Service) checkMenusInitialized(ctx context.Context) error {
	params := &structs.ListMenuParams{}
	count := s.sys.Menu.CountX(ctx, params)
	if count > 0 {
		logger.Infof(ctx, "Menus already exist, synchronizing default menu definitions")
		return s.syncDefaultMenus(ctx)
	}

	return s.initMenus(ctx)
}

// getMenuData returns menu data structure
func (s *Service) getMenuData() *struct {
	Headers  []structs.MenuBody
	Sidebars []structs.MenuBody
	Submenus []structs.MenuBody
	Accounts []structs.MenuBody
	Spaces   []structs.MenuBody
} {
	return &menuData.SystemDefaultMenus
}

// verifyMenuData validates menu data before initialization
func (s *Service) verifyMenuData(ctx context.Context) error {
	menus := s.getMenuData()

	if len(menus.Headers) == 0 {
		return fmt.Errorf("no header menus defined")
	}

	requiredHeaders := map[string]bool{
		"dashboard": false,
		"system":    false,
	}

	for _, header := range menus.Headers {
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

// initMenus initializes default menu structure and creates space relationships
func (s *Service) initMenus(ctx context.Context) error {
	logger.Infof(ctx, "Initializing default menus...")

	if err := s.verifyMenuData(ctx); err != nil {
		return fmt.Errorf("menu data verification failed: %w", err)
	}

	// Get default space
	space, err := s.getDefaultSpace(ctx)
	if err != nil {
		logger.Errorf(ctx, "initMenus error on get default space: %v", err)
		return fmt.Errorf("failed to get default space: %w", err)
	}

	adminUser, err := s.getAdminUser(ctx, "menu creation")
	if err != nil {
		return err
	}

	defaultExtras := make(types.JSON)
	menuData := s.getMenuData()

	var createdMenus, relationshipCount int

	// Create header menus and map IDs
	headerIDMap := make(map[string]string)
	for _, header := range menuData.Headers {
		menuBody := header
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

		// Create space-menu relationship
		_, err = s.ts.SpaceMenu.AddMenuToSpace(ctx, space.ID, createdMenu.ID)
		if err != nil {
			logger.Errorf(ctx, "Error linking menu %s to space %s: %v", createdMenu.ID, space.ID, err)
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

		// Create space-menu relationship
		_, err = s.ts.SpaceMenu.AddMenuToSpace(ctx, space.ID, createdMenu.ID)
		if err != nil {
			logger.Errorf(ctx, "Error linking menu %s to space %s: %v", createdMenu.ID, space.ID, err)
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

		// Create space-menu relationship
		_, err = s.ts.SpaceMenu.AddMenuToSpace(ctx, space.ID, createdMenu.ID)
		if err != nil {
			logger.Errorf(ctx, "Error linking menu %s to space %s: %v", createdMenu.ID, space.ID, err)
			return err
		}
		relationshipCount++
	}

	// Create account menus
	for _, menu := range menuData.Accounts {
		menuBody := menu
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

		// Create space-menu relationship
		_, err = s.ts.SpaceMenu.AddMenuToSpace(ctx, space.ID, createdMenu.ID)
		if err != nil {
			logger.Errorf(ctx, "Error linking menu %s to space %s: %v", createdMenu.ID, space.ID, err)
			return err
		}
		relationshipCount++
	}

	// Create space menus
	for _, menu := range menuData.Spaces {
		menuBody := menu
		menuBody.CreatedBy = &adminUser.ID
		menuBody.UpdatedBy = &adminUser.ID
		menuBody.Extras = &defaultExtras

		logger.Debugf(ctx, "Creating space menu: %s", menuBody.Name)
		createdMenu, err := s.sys.Menu.Create(ctx, &menuBody)
		if err != nil {
			logger.Errorf(ctx, "Error creating space menu %s: %v", menuBody.Name, err)
			return fmt.Errorf("failed to create space menu '%s': %w", menuBody.Name, err)
		}
		createdMenus++

		// Create space-menu relationship
		_, err = s.ts.SpaceMenu.AddMenuToSpace(ctx, space.ID, createdMenu.ID)
		if err != nil {
			logger.Errorf(ctx, "Error linking menu %s to space %s: %v", createdMenu.ID, space.ID, err)
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

type menuSyncIndex struct {
	byKey          map[string]*structs.ReadMenu
	idBySlug       map[string]string
	parentSlugByID map[string]string
}

type defaultMenuSection struct {
	name  string
	menus []structs.MenuBody
}

func defaultMenuSections(menus *struct {
	Headers  []structs.MenuBody
	Sidebars []structs.MenuBody
	Submenus []structs.MenuBody
	Accounts []structs.MenuBody
	Spaces   []structs.MenuBody
}) []defaultMenuSection {
	return []defaultMenuSection{
		{name: "headers", menus: menus.Headers},
		{name: "sidebars", menus: menus.Sidebars},
		{name: "submenus", menus: menus.Submenus},
		{name: "accounts", menus: menus.Accounts},
		{name: "spaces", menus: menus.Spaces},
	}
}

// syncDefaultMenus repairs existing installations without replacing custom menus.
func (s *Service) syncDefaultMenus(ctx context.Context) error {
	logger.Infof(ctx, "Synchronizing default menus...")

	if err := s.verifyMenuData(ctx); err != nil {
		return fmt.Errorf("menu data verification failed: %w", err)
	}

	space, err := s.getDefaultSpace(ctx)
	if err != nil {
		return fmt.Errorf("failed to get default space: %w", err)
	}

	adminUser, err := s.getAdminUser(ctx, "menu synchronization")
	if err != nil {
		return err
	}

	existing, err := s.listMenusForSync(ctx)
	if err != nil {
		return err
	}

	index := newMenuSyncIndex(existing)
	defaultExtras := make(types.JSON)
	sections := defaultMenuSections(s.getMenuData())
	var createdMenus, updatedMenus, relationshipCount, skippedMenus int

	for _, section := range sections {
		for _, defaultMenu := range section.menus {
			menuBody := defaultMenu
			if menuBody.ParentID != "" {
				parentID, ok := index.idBySlug[menuBody.ParentID]
				if !ok {
					logger.Warnf(ctx, "Parent menu '%s' not found while synchronizing default %s menu '%s', skipping",
						menuBody.ParentID, section.name, menuDefinitionDisplayName(menuBody))
					skippedMenus++
					continue
				}
				menuBody.ParentID = parentID
			}

			key := defaultMenuDefinitionKey(defaultMenu)
			current := index.byKey[key]
			if current == nil && defaultMenu.Slug != "" {
				if id, ok := index.idBySlug[defaultMenu.Slug]; ok {
					current = index.findByID(id)
				}
			}

			if current == nil {
				menuBody.CreatedBy = &adminUser.ID
				menuBody.UpdatedBy = &adminUser.ID
				menuBody.Extras = &defaultExtras

				created, err := s.sys.Menu.Create(ctx, &menuBody)
				if err != nil {
					return fmt.Errorf("failed to create default menu '%s': %w", menuDefinitionDisplayName(defaultMenu), err)
				}
				index.add(created, defaultMenu)
				current = created
				createdMenus++
				logger.Debugf(ctx, "Created missing default menu '%s'", menuDefinitionDisplayName(defaultMenu))
			} else if updates, changed := defaultMenuUpdates(current, menuBody, adminUser.ID); changed {
				updated, err := s.sys.Menu.Update(ctx, updates)
				if err != nil {
					return fmt.Errorf("failed to update default menu '%s': %w", menuDefinitionDisplayName(defaultMenu), err)
				}
				index.add(updated, defaultMenu)
				current = updated
				updatedMenus++
				logger.Debugf(ctx, "Updated default menu '%s'", menuDefinitionDisplayName(defaultMenu))
			}

			createdRelationship, err := s.ensureDefaultSpaceMenu(ctx, space.ID, current.ID)
			if err != nil {
				return fmt.Errorf("failed to link default menu '%s' to space '%s': %w",
					menuDefinitionDisplayName(defaultMenu), space.Slug, err)
			}
			if createdRelationship {
				relationshipCount++
			}
		}
	}

	logger.Infof(ctx, "Menu synchronization completed, created %d menus, updated %d menus, created %d space-menu relationships, skipped %d menus",
		createdMenus, updatedMenus, relationshipCount, skippedMenus)
	return nil
}

func (s *Service) listMenusForSync(ctx context.Context) ([]*structs.ReadMenu, error) {
	result, err := s.sys.Menu.List(ctx, &structs.ListMenuParams{Children: true, Limit: 10000})
	if err != nil {
		return nil, fmt.Errorf("failed to list menus for synchronization: %w", err)
	}
	return flattenReadMenuTree(result.Items), nil
}

func flattenReadMenuTree(menus []*structs.ReadMenu) []*structs.ReadMenu {
	flattened := make([]*structs.ReadMenu, 0, len(menus))
	var walk func(items []*structs.ReadMenu)
	walk = func(items []*structs.ReadMenu) {
		for _, menu := range items {
			if menu == nil {
				continue
			}
			flattened = append(flattened, menu)
			if len(menu.Children) == 0 {
				continue
			}
			children := make([]*structs.ReadMenu, 0, len(menu.Children))
			for _, child := range menu.Children {
				switch typed := child.(type) {
				case *structs.ReadMenu:
					children = append(children, typed)
				}
			}
			walk(children)
		}
	}
	walk(menus)
	return flattened
}

func newMenuSyncIndex(menus []*structs.ReadMenu) *menuSyncIndex {
	index := &menuSyncIndex{
		byKey:          make(map[string]*structs.ReadMenu),
		idBySlug:       make(map[string]string),
		parentSlugByID: make(map[string]string),
	}

	for _, menu := range menus {
		if menu.Slug != "" {
			index.idBySlug[menu.Slug] = menu.ID
			index.parentSlugByID[menu.ID] = menu.Slug
		}
	}
	for _, menu := range menus {
		key := readMenuDefinitionKey(menu, index.parentSlugByID)
		if key != "" {
			index.byKey[key] = menu
		}
	}
	return index
}

func (i *menuSyncIndex) add(menu *structs.ReadMenu, source structs.MenuBody) {
	if menu == nil {
		return
	}
	if menu.Slug != "" {
		i.idBySlug[menu.Slug] = menu.ID
		i.parentSlugByID[menu.ID] = menu.Slug
	}
	i.byKey[defaultMenuDefinitionKey(source)] = menu
	if key := readMenuDefinitionKey(menu, i.parentSlugByID); key != "" {
		i.byKey[key] = menu
	}
}

func (i *menuSyncIndex) findByID(id string) *structs.ReadMenu {
	for _, menu := range i.byKey {
		if menu.ID == id {
			return menu
		}
	}
	return nil
}

func defaultMenuUpdates(existing *structs.ReadMenu, expected structs.MenuBody, updatedBy string) (*structs.UpdateMenuBody, bool) {
	updates := &structs.UpdateMenuBody{ID: existing.ID}
	changed := false

	if expected.Name != "" && existing.Name != expected.Name {
		updates.Name = expected.Name
		changed = true
	}
	if expected.Label != "" && existing.Label != expected.Label {
		updates.Label = expected.Label
		changed = true
	}
	if expected.Slug != "" && existing.Slug != expected.Slug {
		updates.Slug = expected.Slug
		changed = true
	}
	if expected.Type != "" && existing.Type != expected.Type {
		updates.Type = expected.Type
		changed = true
	}
	if expected.Path != "" && existing.Path != expected.Path {
		updates.Path = expected.Path
		changed = true
	}
	if expected.Target != "" && existing.Target != expected.Target {
		updates.Target = expected.Target
		changed = true
	}
	if expected.Icon != "" && existing.Icon != expected.Icon {
		updates.Icon = expected.Icon
		changed = true
	}
	if expected.Perms != "" && existing.Perms != expected.Perms {
		updates.Perms = expected.Perms
		changed = true
	}
	if expected.Hidden != nil && existing.Hidden != *expected.Hidden {
		updates.Hidden = expected.Hidden
		changed = true
	}
	if expected.Order != nil && existing.Order != *expected.Order {
		updates.Order = expected.Order
		changed = true
	}
	if expected.Disabled != nil && existing.Disabled != *expected.Disabled {
		updates.Disabled = expected.Disabled
		changed = true
	}
	if expected.ParentID != "" && existing.ParentID != expected.ParentID {
		updates.ParentID = expected.ParentID
		changed = true
	}
	if expected.ParentID == "" && existing.ParentID != "" && existing.ParentID != "root" {
		updates.ParentID = "root"
		changed = true
	}
	if changed {
		updates.UpdatedBy = &updatedBy
	}
	return updates, changed
}

func (s *Service) ensureDefaultSpaceMenu(ctx context.Context, spaceID, menuID string) (bool, error) {
	exists, err := s.ts.SpaceMenu.IsMenuInSpace(ctx, spaceID, menuID)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}
	if _, err := s.ts.SpaceMenu.AddMenuToSpace(ctx, spaceID, menuID); err != nil {
		return false, err
	}
	return true, nil
}

func defaultMenuDefinitionKey(menu structs.MenuBody) string {
	if menu.Slug != "" {
		return "slug:" + menu.Slug
	}
	return menuFallbackKey(menu.Type, menu.ParentID, menu.Path, menu.Label, menu.Name, menuOrderValue(menu.Order))
}

func readMenuDefinitionKey(menu *structs.ReadMenu, parentSlugByID map[string]string) string {
	if menu == nil {
		return ""
	}
	if menu.Slug != "" {
		return "slug:" + menu.Slug
	}
	parentSlug := parentSlugByID[menu.ParentID]
	if menu.ParentID == "" || menu.ParentID == "root" {
		parentSlug = ""
	}
	return menuFallbackKey(menu.Type, parentSlug, menu.Path, menu.Label, menu.Name, menu.Order)
}

func menuFallbackKey(menuType, parentSlug, path, label, name string, order int) string {
	return fmt.Sprintf("fallback:%s:%s:%s:%s:%s:%d", menuType, parentSlug, path, label, name, order)
}

func menuOrderValue(order *int) int {
	if order == nil {
		return 0
	}
	return *order
}

func menuDefinitionDisplayName(menu structs.MenuBody) string {
	if menu.Slug != "" {
		return menu.Slug
	}
	if menu.Label != "" {
		return menu.Label
	}
	return menu.Name
}

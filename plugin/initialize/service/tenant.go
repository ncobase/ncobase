package service

import (
	"context"
	"fmt"
	tenantStructs "ncobase/tenant/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// checkTenantsInitialized checks if tenants are already initialized
func (s *Service) checkTenantsInitialized(ctx context.Context) error {
	defaultSlug := s.getDefaultTenantSlug()

	tenant, err := s.ts.Tenant.GetBySlug(ctx, defaultSlug)
	if err == nil && tenant != nil {
		logger.Infof(ctx, "Default tenant already exists, skipping initialization")
		return nil
	}

	params := &tenantStructs.ListTenantParams{}
	count := s.ts.Tenant.CountX(ctx, params)
	if count > 0 {
		logger.Infof(ctx, "Tenants already exist, skipping initialization")
		return nil
	}

	return s.initTenants(ctx)
}

// initTenants initializes tenants using current data mode
func (s *Service) initTenants(ctx context.Context) error {
	logger.Infof(ctx, "Initializing system tenants in %s mode...", s.state.DataMode)

	dataLoader := s.getDataLoader()
	tenants := dataLoader.GetTenants()

	var createdCount int
	var tenantMap = make(map[string]string) // slug -> id mapping

	// Step 1: Create tenants
	for _, tenant := range tenants {
		existing, err := s.ts.Tenant.GetBySlug(ctx, tenant.Slug)
		if err == nil && existing != nil {
			logger.Infof(ctx, "Tenant %s already exists, skipping", tenant.Slug)
			tenantMap[tenant.Slug] = existing.ID
			continue
		}

		created, err := s.ts.Tenant.Create(ctx, &tenant)
		if err != nil {
			logger.Errorf(ctx, "Error creating tenant %s: %v", tenant.Name, err)
			return fmt.Errorf("failed to create tenant '%s': %w", tenant.Name, err)
		}
		logger.Debugf(ctx, "Created tenant: %s", tenant.Name)
		tenantMap[tenant.Slug] = created.ID
		createdCount++
	}

	// Step 2: Initialize tenant quotas
	if err := s.initTenantQuotas(ctx, dataLoader, tenantMap); err != nil {
		return fmt.Errorf("failed to initialize tenant quotas: %w", err)
	}

	// Step 3: Initialize tenant settings
	if err := s.initTenantSettings(ctx, dataLoader, tenantMap); err != nil {
		return fmt.Errorf("failed to initialize tenant settings: %w", err)
	}

	if createdCount == 0 {
		logger.Warnf(ctx, "No tenants were created during initialization")
	}

	// Verify default tenant exists
	defaultSlug := s.getDefaultTenantSlug()
	defaultTenant, err := s.ts.Tenant.GetBySlug(ctx, defaultSlug)
	if err != nil || defaultTenant == nil {
		logger.Errorf(ctx, "Default tenant '%s' does not exist after initialization", defaultSlug)
		return fmt.Errorf("default tenant '%s' not found after initialization: %w", defaultSlug, err)
	}

	count := s.ts.Tenant.CountX(ctx, &tenantStructs.ListTenantParams{})
	logger.Infof(ctx, "Tenant initialization completed in %s mode, total %d tenants", s.state.DataMode, count)
	return nil
}

// initTenantQuotas initializes tenant quotas for all tenants
func (s *Service) initTenantQuotas(ctx context.Context, dataLoader DataLoader, tenantMap map[string]string) error {
	logger.Infof(ctx, "Initializing tenant quotas...")

	quotaTemplates := dataLoader.GetTenantQuotas()
	if len(quotaTemplates) == 0 {
		logger.Warnf(ctx, "No quota templates found, skipping quota initialization")
		return nil
	}

	var createdCount int
	for tenantSlug, tenantID := range tenantMap {
		for _, quotaTemplate := range quotaTemplates {
			// Check if quota already exists
			existing, err := s.ts.TenantQuota.GetByTenantAndType(ctx, tenantID, quotaTemplate.QuotaType)
			if err == nil && existing != nil {
				logger.Debugf(ctx, "Quota %s already exists for tenant %s, skipping", quotaTemplate.QuotaType, tenantSlug)
				continue
			}

			// Create quota for this tenant
			quota := quotaTemplate
			quota.TenantID = tenantID

			if _, err := s.ts.TenantQuota.Create(ctx, &quota); err != nil {
				logger.Errorf(ctx, "Error creating quota %s for tenant %s: %v", quota.QuotaType, tenantSlug, err)
				return fmt.Errorf("failed to create quota '%s' for tenant '%s': %w", quota.QuotaType, tenantSlug, err)
			}
			logger.Debugf(ctx, "Created quota %s for tenant %s", quota.QuotaType, tenantSlug)
			createdCount++
		}
	}

	logger.Infof(ctx, "Created %d tenant quotas", createdCount)
	return nil
}

// initTenantSettings initializes tenant settings for all tenants
func (s *Service) initTenantSettings(ctx context.Context, dataLoader DataLoader, tenantMap map[string]string) error {
	logger.Infof(ctx, "Initializing tenant settings...")

	settingTemplates := dataLoader.GetTenantSettings()
	if len(settingTemplates) == 0 {
		logger.Warnf(ctx, "No setting templates found, skipping settings initialization")
		return nil
	}

	var createdCount int
	for tenantSlug, tenantID := range tenantMap {
		for _, settingTemplate := range settingTemplates {
			// Check if setting already exists
			existing, err := s.ts.TenantSetting.GetByKey(ctx, tenantID, settingTemplate.SettingKey)
			if err == nil && existing != nil {
				logger.Debugf(ctx, "Setting %s already exists for tenant %s, skipping", settingTemplate.SettingKey, tenantSlug)
				continue
			}

			// Create setting for this tenant
			setting := settingTemplate
			setting.TenantID = tenantID

			if _, err := s.ts.TenantSetting.Create(ctx, &setting); err != nil {
				logger.Errorf(ctx, "Error creating setting %s for tenant %s: %v", setting.SettingKey, tenantSlug, err)
				return fmt.Errorf("failed to create setting '%s' for tenant '%s': %w", setting.SettingKey, tenantSlug, err)
			}
			logger.Debugf(ctx, "Created setting %s for tenant %s", setting.SettingKey, tenantSlug)
			createdCount++
		}
	}

	logger.Infof(ctx, "Created %d tenant settings", createdCount)
	return nil
}

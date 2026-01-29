package repository

import (
	"ncobase/core/space/data/ent"
	"ncobase/core/space/structs"
)

// SerializeSpace converts ent.Space to structs.ReadSpace.
func SerializeSpace(row *ent.Space) *structs.ReadSpace {
	if row == nil {
		return nil
	}
	return &structs.ReadSpace{
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

// SerializeSpaces converts ent.Space list to structs.ReadSpace list.
func SerializeSpaces(rows []*ent.Space) []*structs.ReadSpace {
	result := make([]*structs.ReadSpace, 0, len(rows))
	for _, row := range rows {
		result = append(result, SerializeSpace(row))
	}
	return result
}

// SerializeSpaceOption converts ent.SpaceOption to structs.SpaceOption.
func SerializeSpaceOption(row *ent.SpaceOption) *structs.SpaceOption {
	if row == nil {
		return nil
	}
	return &structs.SpaceOption{
		SpaceID:  row.SpaceID,
		OptionID: row.OptionID,
	}
}

// SerializeSpaceMenu converts ent.SpaceMenu to structs.SpaceMenu.
func SerializeSpaceMenu(row *ent.SpaceMenu) *structs.SpaceMenu {
	if row == nil {
		return nil
	}
	return &structs.SpaceMenu{
		SpaceID: row.SpaceID,
		MenuID:  row.MenuID,
	}
}

// SerializeSpaceDictionary converts ent.SpaceDictionary to structs.SpaceDictionary.
func SerializeSpaceDictionary(row *ent.SpaceDictionary) *structs.SpaceDictionary {
	if row == nil {
		return nil
	}
	return &structs.SpaceDictionary{
		SpaceID:      row.SpaceID,
		DictionaryID: row.DictionaryID,
	}
}

// SerializeSpaceOrganization converts ent.SpaceOrganization to structs.SpaceOrganization.
func SerializeSpaceOrganization(row *ent.SpaceOrganization) *structs.SpaceOrganization {
	if row == nil {
		return nil
	}
	return &structs.SpaceOrganization{
		SpaceID: row.SpaceID,
		OrgID:   row.OrgID,
	}
}

// SerializeSpaceOrganizations converts ent.SpaceOrganization list to structs.SpaceOrganization list.
func SerializeSpaceOrganizations(rows []*ent.SpaceOrganization) []*structs.SpaceOrganization {
	result := make([]*structs.SpaceOrganization, 0, len(rows))
	for _, row := range rows {
		result = append(result, SerializeSpaceOrganization(row))
	}
	return result
}

// SerializeSpaceQuota converts ent.SpaceQuota to structs.ReadSpaceQuota.
func SerializeSpaceQuota(row *ent.SpaceQuota) *structs.ReadSpaceQuota {
	if row == nil {
		return nil
	}
	result := &structs.ReadSpaceQuota{
		ID:          row.ID,
		SpaceID:     row.SpaceID,
		QuotaType:   structs.QuotaType(row.QuotaType),
		QuotaName:   row.QuotaName,
		MaxValue:    row.MaxValue,
		CurrentUsed: row.CurrentUsed,
		Unit:        structs.QuotaUnit(row.Unit),
		Description: row.Description,
		Enabled:     row.Enabled,
		Extras:      &row.Extras,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}
	result.CalculateUtilization()
	return result
}

// SerializeSpaceQuotas converts ent.SpaceQuota list to structs.ReadSpaceQuota list.
func SerializeSpaceQuotas(rows []*ent.SpaceQuota) []*structs.ReadSpaceQuota {
	result := make([]*structs.ReadSpaceQuota, 0, len(rows))
	for _, row := range rows {
		result = append(result, SerializeSpaceQuota(row))
	}
	return result
}

// SerializeSpaceBilling converts ent.SpaceBilling to structs.ReadSpaceBilling.
func SerializeSpaceBilling(row *ent.SpaceBilling) *structs.ReadSpaceBilling {
	if row == nil {
		return nil
	}
	result := &structs.ReadSpaceBilling{
		ID:            row.ID,
		SpaceID:       row.SpaceID,
		BillingPeriod: structs.BillingPeriod(row.BillingPeriod),
		PeriodStart:   &row.PeriodStart,
		PeriodEnd:     &row.PeriodEnd,
		Amount:        row.Amount,
		Currency:      row.Currency,
		Status:        structs.BillingStatus(row.Status),
		Description:   row.Description,
		InvoiceNumber: row.InvoiceNumber,
		PaymentMethod: row.PaymentMethod,
		PaidAt:        &row.PaidAt,
		DueDate:       &row.DueDate,
		UsageDetails:  &row.UsageDetails,
		Extras:        &row.Extras,
		CreatedBy:     &row.CreatedBy,
		CreatedAt:     &row.CreatedAt,
		UpdatedBy:     &row.UpdatedBy,
		UpdatedAt:     &row.UpdatedAt,
	}
	result.CalculateOverdue()
	return result
}

// SerializeSpaceBillings converts ent.SpaceBilling list to structs.ReadSpaceBilling list.
func SerializeSpaceBillings(rows []*ent.SpaceBilling) []*structs.ReadSpaceBilling {
	result := make([]*structs.ReadSpaceBilling, 0, len(rows))
	for _, row := range rows {
		result = append(result, SerializeSpaceBilling(row))
	}
	return result
}

// SerializeSpaceSetting converts ent.SpaceSetting to structs.ReadSpaceSetting.
func SerializeSpaceSetting(row *ent.SpaceSetting) *structs.ReadSpaceSetting {
	if row == nil {
		return nil
	}
	return &structs.ReadSpaceSetting{
		ID:           row.ID,
		SpaceID:      row.SpaceID,
		SettingKey:   row.SettingKey,
		SettingName:  row.SettingName,
		SettingValue: row.SettingValue,
		DefaultValue: row.DefaultValue,
		SettingType:  structs.SettingType(row.SettingType),
		Scope:        structs.SettingScope(row.Scope),
		Category:     row.Category,
		Description:  row.Description,
		IsPublic:     row.IsPublic,
		IsRequired:   row.IsRequired,
		IsReadonly:   row.IsReadonly,
		Validation:   &row.Validation,
		Extras:       &row.Extras,
		CreatedBy:    &row.CreatedBy,
		CreatedAt:    &row.CreatedAt,
		UpdatedBy:    &row.UpdatedBy,
		UpdatedAt:    &row.UpdatedAt,
	}
}

// SerializeSpaceSettings converts ent.SpaceSetting list to structs.ReadSpaceSetting list.
func SerializeSpaceSettings(rows []*ent.SpaceSetting) []*structs.ReadSpaceSetting {
	result := make([]*structs.ReadSpaceSetting, 0, len(rows))
	for _, row := range rows {
		result = append(result, SerializeSpaceSetting(row))
	}
	return result
}

// SerializeUserSpace converts ent.UserSpace to structs.UserSpace.
func SerializeUserSpace(row *ent.UserSpace) *structs.UserSpace {
	if row == nil {
		return nil
	}
	return &structs.UserSpace{
		UserID:  row.UserID,
		SpaceID: row.SpaceID,
	}
}

// SerializeUserSpaces converts ent.UserSpace list to structs.UserSpace list.
func SerializeUserSpaces(rows []*ent.UserSpace) []*structs.UserSpace {
	result := make([]*structs.UserSpace, 0, len(rows))
	for _, row := range rows {
		result = append(result, SerializeUserSpace(row))
	}
	return result
}

// SerializeUserSpaceRole converts ent.UserSpaceRole to structs.UserSpaceRole.
func SerializeUserSpaceRole(row *ent.UserSpaceRole) *structs.UserSpaceRole {
	if row == nil {
		return nil
	}
	return &structs.UserSpaceRole{
		UserID:  row.UserID,
		SpaceID: row.SpaceID,
		RoleID:  row.RoleID,
	}
}

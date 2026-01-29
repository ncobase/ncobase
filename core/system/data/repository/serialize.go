package repository

import (
	"ncobase/core/system/data/ent"
	"ncobase/core/system/structs"
)

// SerializeDictionary converts ent.Dictionary to structs.ReadDictionary.
func SerializeDictionary(row *ent.Dictionary) *structs.ReadDictionary {
	if row == nil {
		return nil
	}
	return &structs.ReadDictionary{
		ID:        row.ID,
		Name:      row.Name,
		Slug:      row.Slug,
		Type:      row.Type,
		Value:     row.Value,
		CreatedBy: &row.CreatedBy,
		CreatedAt: &row.CreatedAt,
		UpdatedBy: &row.UpdatedBy,
		UpdatedAt: &row.UpdatedAt,
	}
}

// SerializeDictionaries converts []*ent.Dictionary to []*structs.ReadDictionary.
func SerializeDictionaries(rows []*ent.Dictionary) []*structs.ReadDictionary {
	rs := make([]*structs.ReadDictionary, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, SerializeDictionary(row))
	}
	return rs
}

// SerializeMenu converts ent.Menu to structs.ReadMenu.
func SerializeMenu(row *ent.Menu) *structs.ReadMenu {
	if row == nil {
		return nil
	}
	return &structs.ReadMenu{
		ID:        row.ID,
		Name:      row.Name,
		Label:     row.Label,
		Slug:      row.Slug,
		Type:      row.Type,
		Path:      row.Path,
		Target:    row.Target,
		Icon:      row.Icon,
		Perms:     row.Perms,
		Hidden:    row.Hidden,
		Order:     row.Order,
		Disabled:  row.Disabled,
		Extras:    &row.Extras,
		ParentID:  row.ParentID,
		CreatedBy: &row.CreatedBy,
		CreatedAt: &row.CreatedAt,
		UpdatedBy: &row.UpdatedBy,
		UpdatedAt: &row.UpdatedAt,
	}
}

// SerializeMenus converts []*ent.Menu to []*structs.ReadMenu.
func SerializeMenus(rows []*ent.Menu) []*structs.ReadMenu {
	rs := make([]*structs.ReadMenu, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, SerializeMenu(row))
	}
	return rs
}

// SerializeOption converts ent.Options to structs.ReadOption.
func SerializeOption(row *ent.Options) *structs.ReadOption {
	if row == nil {
		return nil
	}
	return &structs.ReadOption{
		ID:        row.ID,
		Name:      row.Name,
		Type:      row.Type,
		Value:     row.Value,
		Autoload:  row.Autoload,
		CreatedBy: &row.CreatedBy,
		CreatedAt: &row.CreatedAt,
		UpdatedBy: &row.UpdatedBy,
		UpdatedAt: &row.UpdatedAt,
	}
}

// SerializeOptions converts []*ent.Options to []*structs.ReadOption.
func SerializeOptions(rows []*ent.Options) []*structs.ReadOption {
	rs := make([]*structs.ReadOption, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, SerializeOption(row))
	}
	return rs
}

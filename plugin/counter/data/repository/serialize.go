package repository

import (
	"ncobase/plugin/counter/data/ent"
	"ncobase/plugin/counter/structs"
)

// SerializeCounter converts ent.Counter to structs.ReadCounter.
func SerializeCounter(row *ent.Counter) *structs.ReadCounter {
	if row == nil {
		return nil
	}
	return &structs.ReadCounter{
		ID:            row.ID,
		Identifier:    row.Identifier,
		Name:          row.Name,
		Prefix:        row.Prefix,
		Suffix:        row.Suffix,
		StartValue:    row.StartValue,
		IncrementStep: row.IncrementStep,
		DateFormat:    row.DateFormat,
		CurrentValue:  row.CurrentValue,
		Disabled:      row.Disabled,
		Description:   row.Description,
		SpaceID:       &row.SpaceID,
		CreatedBy:     &row.CreatedBy,
		CreatedAt:     &row.CreatedAt,
		UpdatedBy:     &row.UpdatedBy,
		UpdatedAt:     &row.UpdatedAt,
	}
}

// SerializeCounters converts []*ent.Counter to []*structs.ReadCounter.
func SerializeCounters(rows []*ent.Counter) []*structs.ReadCounter {
	rs := make([]*structs.ReadCounter, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, SerializeCounter(row))
	}
	return rs
}

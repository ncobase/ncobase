package repository

import (
	"ncobase/core/auth/data/ent"
	"ncobase/core/auth/structs"
)

// SerializeSession converts ent.Session to structs.ReadSession.
func SerializeSession(row *ent.Session) *structs.ReadSession {
	if row == nil {
		return nil
	}
	return &structs.ReadSession{
		ID:           row.ID,
		UserID:       row.UserID,
		TokenID:      row.TokenID,
		DeviceInfo:   &row.DeviceInfo,
		IPAddress:    row.IPAddress,
		UserAgent:    row.UserAgent,
		Location:     row.Location,
		LoginMethod:  row.LoginMethod,
		IsActive:     row.IsActive,
		LastAccessAt: &row.LastAccessAt,
		ExpiresAt:    &row.ExpiresAt,
		CreatedAt:    &row.CreatedAt,
		UpdatedAt:    &row.UpdatedAt,
	}
}

// SerializeSessions converts multiple ent.Session to structs.ReadSession.
func SerializeSessions(rows []*ent.Session) []*structs.ReadSession {
	rs := make([]*structs.ReadSession, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, SerializeSession(row))
	}
	return rs
}

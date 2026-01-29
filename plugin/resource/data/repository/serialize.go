package repository

import (
	"fmt"
	"ncobase/plugin/resource/data/ent"
	"ncobase/plugin/resource/structs"
	"time"

	"github.com/ncobase/ncore/types"
)

// CloneExtras returns a safe copy of extras.
func CloneExtras(extras types.JSON) types.JSON {
	if extras == nil {
		return make(types.JSON)
	}
	cloned := make(types.JSON)
	for k, v := range extras {
		cloned[k] = v
	}
	return cloned
}

// CloneExtrasPtr returns a safe copy of extras from a pointer.
func CloneExtrasPtr(extras *types.JSON) types.JSON {
	if extras == nil {
		return make(types.JSON)
	}
	return CloneExtras(*extras)
}

// SerializeFile converts ent.File to structs.ReadFile.
func SerializeFile(row *ent.File) *structs.ReadFile {
	if row == nil {
		return nil
	}

	extras := CloneExtras(row.Extras)

	file := &structs.ReadFile{
		ID:           row.ID,
		Name:         row.Name,
		OriginalName: row.OriginalName,
		Path:         row.Path,
		Type:         row.Type,
		Size:         &row.Size,
		Storage:      row.Storage,
		Bucket:       row.Bucket,
		Endpoint:     row.Endpoint,
		AccessLevel:  structs.AccessLevel(row.AccessLevel),
		ExpiresAt:    row.ExpiresAt,
		Tags:         row.Tags,
		IsPublic:     row.IsPublic,
		Category:     structs.FileCategory(row.Category),
		Hash:         row.Hash,
		OwnerID:      row.OwnerID,
		Extras:       &row.Extras,
		CreatedBy:    &row.CreatedBy,
		CreatedAt:    &row.CreatedAt,
		UpdatedBy:    &row.UpdatedBy,
		UpdatedAt:    &row.UpdatedAt,
	}

	if file.ExpiresAt != nil {
		file.IsExpired = time.Now().UnixMilli() > *file.ExpiresAt
	}

	if file.IsPublic {
		file.DownloadURL = fmt.Sprintf("/res/dl/%s", row.ID)
		if thumbnailPath, ok := extras["thumbnail_path"].(string); ok && thumbnailPath != "" {
			file.ThumbnailURL = fmt.Sprintf("/res/thumb/%s", row.ID)
		}
	}

	return file
}

// SerializeFiles converts ent.File list to structs.ReadFile list.
func SerializeFiles(rows []*ent.File) []*structs.ReadFile {
	results := make([]*structs.ReadFile, 0, len(rows))
	for _, row := range rows {
		results = append(results, SerializeFile(row))
	}
	return results
}

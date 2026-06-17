package repository

import (
	"ncobase/plugin/ai/data/ent"
	"ncobase/plugin/ai/structs"
)

func CloneMetadata(metadata map[string]any) map[string]any {
	if metadata == nil {
		return nil
	}
	cloned := make(map[string]any, len(metadata))
	for key, value := range metadata {
		cloned[key] = value
	}
	return cloned
}

func SerializeRun(row *ent.AIRun) *structs.Run {
	if row == nil {
		return nil
	}

	return &structs.Run{
		ID:                 row.ID,
		OperationID:        row.OperationID,
		Action:             row.Action,
		Mode:               structs.RunMode(row.Mode),
		Status:             structs.RunStatus(row.Status),
		Provider:           row.Provider,
		Model:              row.Model,
		FallbackModel:      row.FallbackModel,
		InputTokens:        row.InputTokens,
		OutputTokens:       row.OutputTokens,
		TotalTokens:        row.TotalTokens,
		ReasoningTokens:    row.ReasoningTokens,
		CacheCreatedTokens: row.CacheCreatedTokens,
		CacheReadTokens:    row.CacheReadTokens,
		DurationMS:         row.DurationMs,
		ErrorCode:          row.ErrorCode,
		ErrorMessage:       row.ErrorMessage,
		RequestHash:        row.RequestHash,
		ResponseHash:       row.ResponseHash,
		EstimatedCost:      row.EstimatedCost,
		Currency:           row.Currency,
		Metadata:           CloneMetadata(row.Metadata),
		SpaceID:            row.SpaceID,
		UserID:             row.UserID,
		CreatedBy:          row.CreatedBy,
		UpdatedBy:          row.UpdatedBy,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}

func SerializeRuns(rows []*ent.AIRun) []*structs.Run {
	results := make([]*structs.Run, 0, len(rows))
	for _, row := range rows {
		results = append(results, SerializeRun(row))
	}
	return results
}

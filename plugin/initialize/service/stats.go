package service

import (
	"context"
	"encoding/json"
	"fmt"
	systemStructs "ncobase/system/structs"
	"time"

	"github.com/ncobase/ncore/logging/logger"
)

// initBackup holds initialization backup data
type initBackup struct {
	Timestamp int64     `json:"timestamp"`
	State     InitState `json:"state"`
}

// BackupState creates a backup of the current initialization state
func (s *Service) BackupState(ctx context.Context) error {
	if s.sys == nil {
		return fmt.Errorf("system service not initialized")
	}

	backup := initBackup{
		Timestamp: time.Now().UnixMilli(),
		State:     *s.state,
	}

	// Store backup in options
	backupJSON, err := json.Marshal(backup)
	if err != nil {
		return fmt.Errorf("failed to marshal backup state: %w", err)
	}

	backupKey := fmt.Sprintf("system.initialization.backup.%d", backup.Timestamp)
	createBody := &systemStructs.OptionBody{
		Name:     backupKey,
		Type:     "json",
		Value:    string(backupJSON),
		Autoload: false,
	}

	_, err = s.sys.Option.Create(ctx, createBody)
	if err != nil {
		return fmt.Errorf("failed to create state backup: %w", err)
	}

	logger.Infof(ctx, "Created initialization state backup at timestamp %d", backup.Timestamp)
	return nil
}

// ListBackups lists available initialization state backups
func (s *Service) ListBackups(ctx context.Context) ([]initBackup, error) {
	if s.sys == nil {
		return nil, fmt.Errorf("system service not initialized")
	}

	// Find all backup options
	params := &systemStructs.ListOptionParams{
		Prefix: "system.initialization.backup.",
	}

	options, err := s.sys.Option.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list backup options: %w", err)
	}

	backups := make([]initBackup, 0, len(options.Items))
	for _, option := range options.Items {
		var backup initBackup
		if err := json.Unmarshal([]byte(option.Value), &backup); err != nil {
			logger.Warnf(ctx, "Failed to parse backup %s: %v", option.Name, err)
			continue
		}
		backups = append(backups, backup)
	}

	return backups, nil
}

// RestoreBackup restores initialization state from a backup
func (s *Service) RestoreBackup(ctx context.Context, timestamp int64) error {
	if s.sys == nil {
		return fmt.Errorf("system service not initialized")
	}

	backupKey := fmt.Sprintf("system.initialization.backup.%d", timestamp)
	option, err := s.sys.Option.GetByName(ctx, backupKey)
	if err != nil || option == nil {
		return fmt.Errorf("backup with timestamp %d not found", timestamp)
	}

	var backup initBackup
	if err := json.Unmarshal([]byte(option.Value), &backup); err != nil {
		return fmt.Errorf("failed to parse backup: %w", err)
	}

	// Restore state
	s.state = &backup.State

	// Update current state
	if err := s.SaveState(ctx); err != nil {
		return fmt.Errorf("failed to save restored state: %w", err)
	}

	logger.Infof(ctx, "Restored initialization state from backup timestamp %d", timestamp)
	return nil
}

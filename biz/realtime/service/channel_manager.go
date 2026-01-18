package service

import (
	"context"
	"errors"
	"fmt"

	"ncobase/biz/realtime/structs"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// ChannelManagerService coordinates event-driven channel provisioning.
type ChannelManagerService interface {
	RegisterChannelManagers(em ext.ManagerInterface)
	EnsureChannel(ctx context.Context, channelID, channelType, name string) error
}

type channelManagerService struct {
	channelService ChannelService
}

// NewChannelManagerService creates a channel manager service.
func NewChannelManagerService(channelService ChannelService) ChannelManagerService {
	return &channelManagerService{
		channelService: channelService,
	}
}

func (s *channelManagerService) RegisterChannelManagers(em ext.ManagerInterface) {
	if em == nil {
		logger.Warnf(context.Background(), "Extension manager is nil, cannot register channel managers")
		return
	}

	// Register custom event subscriptions here when new domains need auto-channels.
}

func (s *channelManagerService) EnsureChannel(ctx context.Context, channelID, channelType, name string) error {
	if channelID == "" {
		return errors.New("channel_id is required")
	}

	if _, err := s.channelService.Get(ctx, &structs.FindChannel{ID: channelID}); err == nil {
		return nil
	}

	_, err := s.channelService.Create(ctx, &structs.CreateChannel{
		Channel: structs.ChannelBody{
			Name:        name,
			Description: fmt.Sprintf("Auto-created channel for %s", channelType),
			Type:        channelType,
			Status:      1,
		},
	})

	if err != nil {
		logger.Errorf(ctx, "Failed to create channel %s: %v", channelID, err)
		return err
	}

	logger.Infof(ctx, "Created channel %s for %s", channelID, channelType)
	return nil
}

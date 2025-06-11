package service

import (
	"ncobase/content/data"
	"ncobase/content/wrapper"

	ext "github.com/ncobase/ncore/extension/types"
)

// Service represents content service
type Service struct {
	Taxonomy     TaxonomyServiceInterface
	Topic        TopicServiceInterface
	Channel      ChannelServiceInterface
	Distribution DistributionServiceInterface
	Media        MediaServiceInterface
	TopicMedia   TopicMediaServiceInterface
	rsw          *wrapper.ResourceServiceWrapper
}

// New creates new service
func New(em ext.ManagerInterface, d *data.Data) *Service {
	// Create resource service wrapper
	rsw := wrapper.NewResourceServiceWrapper(em)

	// Create services
	ts := NewTaxonomyService(d)
	tops := NewTopicService(d, ts)
	cs := NewChannelService(d)
	ds := NewDistributionService(d, tops, cs)
	ms := NewMediaService(d, rsw)
	tms := NewTopicMediaService(d)

	return &Service{
		Taxonomy:     ts,
		Topic:        tops,
		Channel:      cs,
		Distribution: ds,
		Media:        ms,
		TopicMedia:   tms,
		rsw:          rsw,
	}
}

// RefreshDependencies refreshes external service dependencies
func (s *Service) RefreshDependencies() {
	s.rsw.RefreshServices()
}

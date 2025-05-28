package service

import "ncobase/content/data"

// Service represents the content service.
type Service struct {
	Taxonomy         TaxonomyServiceInterface
	TaxonomyRelation TaxonomyRelationServiceInterface
	Topic            TopicServiceInterface
	Channel          ChannelServiceInterface
	Distribution     DistributionServiceInterface
	Media            MediaServiceInterface
	TopicMedia       TopicMediaServiceInterface
}

// New creates a new service.
func New(d *data.Data) *Service {
	return &Service{
		Taxonomy:         NewTaxonomyService(d),
		TaxonomyRelation: NewTaxonomyRelationService(d),
		Topic:            NewTopicService(d),
		Channel:          NewChannelService(d),
		Distribution:     NewDistributionService(d),
		Media:            NewMediaService(d),
		TopicMedia:       NewTopicMediaService(d),
	}
}

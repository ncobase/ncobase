package repository

import (
	"ncobase/biz/content/data/ent"
	"ncobase/biz/content/structs"

	"github.com/ncobase/ncore/types"
)

func stringExtra(extras types.JSON, key string) string {
	if value, ok := extras[key]; ok {
		if v, ok := value.(string); ok {
			return v
		}
	}
	return ""
}

func intExtra(extras types.JSON, key string) *int {
	if value, ok := extras[key]; ok {
		switch v := value.(type) {
		case int:
			return &v
		case int64:
			i := int(v)
			return &i
		case float64:
			i := int(v)
			return &i
		}
	}
	return nil
}

func floatExtra(extras types.JSON, key string) *float64 {
	if value, ok := extras[key]; ok {
		switch v := value.(type) {
		case float64:
			return &v
		case int:
			f := float64(v)
			return &f
		case int64:
			f := float64(v)
			return &f
		}
	}
	return nil
}

// SerializeTaxonomy converts ent.Taxonomy to structs.ReadTaxonomy.
func SerializeTaxonomy(row *ent.Taxonomy) *structs.ReadTaxonomy {
	if row == nil {
		return nil
	}
	return &structs.ReadTaxonomy{
		ID:          row.ID,
		Name:        row.Name,
		Type:        row.Type,
		Slug:        row.Slug,
		Cover:       row.Cover,
		Thumbnail:   row.Thumbnail,
		Color:       row.Color,
		Icon:        row.Icon,
		URL:         row.URL,
		Keywords:    row.Keywords,
		Description: row.Description,
		Status:      row.Status,
		Extras:      &row.Extras,
		ParentID:    &row.ParentID,
		SpaceID:     row.ParentID,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}
}

// SerializeTaxonomies converts ent.Taxonomy list to structs.ReadTaxonomy list.
func SerializeTaxonomies(rows []*ent.Taxonomy) []*structs.ReadTaxonomy {
	result := make([]*structs.ReadTaxonomy, 0, len(rows))
	for _, row := range rows {
		result = append(result, SerializeTaxonomy(row))
	}
	return result
}

// SerializeTopic converts ent.Topic to structs.ReadTopic, including extras metadata.
func SerializeTopic(row *ent.Topic) *structs.ReadTopic {
	if row == nil {
		return nil
	}
	result := &structs.ReadTopic{
		ID:         row.ID,
		Name:       row.Name,
		Title:      row.Title,
		Slug:       row.Slug,
		Content:    row.Content,
		Thumbnail:  row.Thumbnail,
		Temp:       row.Temp,
		Markdown:   row.Markdown,
		Private:    row.Private,
		Status:     row.Status,
		Released:   row.Released,
		TaxonomyID: row.TaxonomyID,
		SpaceID:    row.SpaceID,
		CreatedBy:  &row.CreatedBy,
		CreatedAt:  &row.CreatedAt,
		UpdatedBy:  &row.UpdatedBy,
		UpdatedAt:  &row.UpdatedAt,
	}

	if row.Extras != nil {
		if version, ok := row.Extras["version"].(float64); ok {
			result.Version = int(version)
		}
		if contentType, ok := row.Extras["content_type"].(string); ok {
			result.ContentType = contentType
		}
		if seoTitle, ok := row.Extras["seo_title"].(string); ok {
			result.SEOTitle = seoTitle
		}
		if seoDescription, ok := row.Extras["seo_description"].(string); ok {
			result.SEODescription = seoDescription
		}
		if seoKeywords, ok := row.Extras["seo_keywords"].(string); ok {
			result.SEOKeywords = seoKeywords
		}
		if excerptAuto, ok := row.Extras["excerpt_auto"].(bool); ok {
			result.ExcerptAuto = excerptAuto
		}
		if excerpt, ok := row.Extras["excerpt"].(string); ok {
			result.Excerpt = excerpt
		}
		if featuredMedia, ok := row.Extras["featured_media"].(string); ok {
			result.FeaturedMedia = featuredMedia
		}
		if tags, ok := row.Extras["tags"].([]any); ok {
			tagStrings := make([]string, 0, len(tags))
			for _, tag := range tags {
				if tagStr, ok := tag.(string); ok {
					tagStrings = append(tagStrings, tagStr)
				}
			}
			result.Tags = tagStrings
		}
		result.Metadata = &row.Extras
	}

	return result
}

// SerializeTopics converts ent.Topic list to structs.ReadTopic list.
func SerializeTopics(rows []*ent.Topic) []*structs.ReadTopic {
	result := make([]*structs.ReadTopic, 0, len(rows))
	for _, row := range rows {
		result = append(result, SerializeTopic(row))
	}
	return result
}

// SerializeMedia converts ent.Media to structs.ReadMedia, including extras metadata.
func SerializeMedia(row *ent.Media) *structs.ReadMedia {
	if row == nil {
		return nil
	}
	result := &structs.ReadMedia{
		ID:          row.ID,
		Title:       row.Title,
		Type:        row.Type,
		ResourceID:  row.ResourceID,
		URL:         row.URL,
		Description: row.Description,
		Alt:         row.Alt,
		SpaceID:     row.SpaceID,
		OwnerID:     row.OwnerID,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}

	if row.Extras != nil {
		if result.ResourceID == "" {
			if resourceID, ok := row.Extras["resource_id"].(string); ok {
				result.ResourceID = resourceID
			}
		}
		if result.Description == "" {
			if description, ok := row.Extras["description"].(string); ok {
				result.Description = description
			}
		}
		if result.Alt == "" {
			if alt, ok := row.Extras["alt"].(string); ok {
				result.Alt = alt
			}
		}
		if result.OwnerID == "" {
			if ownerID, ok := row.Extras["owner_id"].(string); ok {
				result.OwnerID = ownerID
			}
		}
		result.Path = stringExtra(row.Extras, "path")
		result.MimeType = stringExtra(row.Extras, "mime_type")
		result.Size = intExtra(row.Extras, "size")
		result.Width = intExtra(row.Extras, "width")
		result.Height = intExtra(row.Extras, "height")
		result.Duration = floatExtra(row.Extras, "duration")
		result.Metadata = &row.Extras
	}

	return result
}

// SerializeMedias converts ent.Media list to structs.ReadMedia list.
func SerializeMedias(rows []*ent.Media) []*structs.ReadMedia {
	result := make([]*structs.ReadMedia, 0, len(rows))
	for _, row := range rows {
		result = append(result, SerializeMedia(row))
	}
	return result
}

// SerializeChannel converts ent.CMSChannel to structs.ReadChannel.
func SerializeChannel(row *ent.CMSChannel) *structs.ReadChannel {
	if row == nil {
		return nil
	}
	return &structs.ReadChannel{
		ID:            row.ID,
		Name:          row.Name,
		Type:          row.Type,
		Slug:          row.Slug,
		Icon:          row.Icon,
		Status:        row.Status,
		AllowedTypes:  row.AllowedTypes,
		Config:        &row.Extras,
		Description:   row.Description,
		Logo:          row.Logo,
		WebhookURL:    row.WebhookURL,
		AutoPublish:   row.AutoPublish,
		RequireReview: row.RequireReview,
		SpaceID:       row.SpaceID,
		CreatedBy:     &row.CreatedBy,
		CreatedAt:     &row.CreatedAt,
		UpdatedBy:     &row.UpdatedBy,
		UpdatedAt:     &row.UpdatedAt,
	}
}

// SerializeChannels converts ent.CMSChannel list to structs.ReadChannel list.
func SerializeChannels(rows []*ent.CMSChannel) []*structs.ReadChannel {
	result := make([]*structs.ReadChannel, 0, len(rows))
	for _, row := range rows {
		result = append(result, SerializeChannel(row))
	}
	return result
}

// SerializeDistribution converts ent.Distribution to structs.ReadDistribution.
func SerializeDistribution(row *ent.Distribution) *structs.ReadDistribution {
	if row == nil {
		return nil
	}
	return &structs.ReadDistribution{
		ID:           row.ID,
		TopicID:      row.TopicID,
		ChannelID:    row.ChannelID,
		Status:       row.Status,
		ScheduledAt:  row.ScheduledAt,
		PublishedAt:  row.PublishedAt,
		MetaData:     &row.Extras,
		ExternalID:   row.ExternalID,
		ExternalURL:  row.ExternalURL,
		CustomData:   &row.Extras,
		ErrorDetails: row.ErrorDetails,
		SpaceID:      row.SpaceID,
		CreatedBy:    &row.CreatedBy,
		CreatedAt:    &row.CreatedAt,
		UpdatedBy:    &row.UpdatedBy,
		UpdatedAt:    &row.UpdatedAt,
	}
}

// SerializeDistributions converts ent.Distribution list to structs.ReadDistribution list.
func SerializeDistributions(rows []*ent.Distribution) []*structs.ReadDistribution {
	result := make([]*structs.ReadDistribution, 0, len(rows))
	for _, row := range rows {
		result = append(result, SerializeDistribution(row))
	}
	return result
}

// SerializeTopicMedia converts ent.TopicMedia to structs.ReadTopicMedia.
func SerializeTopicMedia(row *ent.TopicMedia) *structs.ReadTopicMedia {
	if row == nil {
		return nil
	}
	return &structs.ReadTopicMedia{
		ID:        row.ID,
		TopicID:   row.TopicID,
		MediaID:   row.MediaID,
		Type:      row.Type,
		Order:     row.Order,
		CreatedBy: &row.CreatedBy,
		CreatedAt: &row.CreatedAt,
		UpdatedBy: &row.UpdatedBy,
		UpdatedAt: &row.UpdatedAt,
	}
}

// SerializeTopicMedias converts ent.TopicMedia list to structs.ReadTopicMedia list.
func SerializeTopicMedias(rows []*ent.TopicMedia) []*structs.ReadTopicMedia {
	result := make([]*structs.ReadTopicMedia, 0, len(rows))
	for _, row := range rows {
		result = append(result, SerializeTopicMedia(row))
	}
	return result
}

// SerializeTaxonomyRelation converts ent.TaxonomyRelation to structs.ReadTaxonomyRelation.
func SerializeTaxonomyRelation(row *ent.TaxonomyRelation) *structs.ReadTaxonomyRelation {
	if row == nil {
		return nil
	}
	return &structs.ReadTaxonomyRelation{
		ID:         row.ID,
		ObjectID:   row.ObjectID,
		TaxonomyID: row.TaxonomyID,
		Type:       row.Type,
		Order:      &row.Order,
		CreatedBy:  &row.CreatedBy,
		CreatedAt:  &row.CreatedAt,
	}
}

// SerializeTaxonomyRelations converts ent.TaxonomyRelation list to structs.ReadTaxonomyRelation list.
func SerializeTaxonomyRelations(rows []*ent.TaxonomyRelation) []*structs.ReadTaxonomyRelation {
	result := make([]*structs.ReadTaxonomyRelation, 0, len(rows))
	for _, row := range rows {
		result = append(result, SerializeTaxonomyRelation(row))
	}
	return result
}

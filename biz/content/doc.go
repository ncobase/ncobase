// Package content provides content management system (CMS) functionality for the Ncobase platform.
//
// This package implements a comprehensive CMS with support for topics, media management,
// taxonomy, channels, and content distribution. It enables creation, organization, and
// publishing of various content types.
//
// Key Features:
//   - Topic management (articles, posts, pages)
//   - Media library and asset management
//   - Taxonomy system (categories, tags)
//   - Content channels and distribution
//   - Content versioning and history
//   - Publishing workflow and scheduling
//   - Content search and filtering
//
// Main Components:
//   - Handler: HTTP endpoints for content operations
//   - Service: Business logic for content management
//   - Repository: Data access layer for content entities
//
// Content Entities:
//   - Topic: Main content items (articles, posts, pages)
//   - Media: Images, videos, documents, and other assets
//   - Taxonomy: Categories, tags, and classification
//   - Channel: Content distribution channels
//   - Comment: User comments and discussions
//
// Usage Example:
//
//	// Create a new topic
//	topic, err := contentService.CreateTopic(ctx, &structs.CreateTopicInput{
//	    Title:       "Getting Started with Ncobase",
//	    Content:     "This is a comprehensive guide...",
//	    Type:        "article",
//	    Status:      "published",
//	    CategoryID:  categoryID,
//	    Tags:        []string{"tutorial", "guide"},
//	})
//	if err != nil {
//	    return err
//	}
//
//	// Upload media
//	media, err := contentService.UploadMedia(ctx, &structs.UploadMediaInput{
//	    File:        fileReader,
//	    Filename:    "hero-image.jpg",
//	    ContentType: "image/jpeg",
//	    Alt:         "Hero image for article",
//	})
//
//	// Create taxonomy
//	category, err := contentService.CreateCategory(ctx, &structs.CreateCategoryInput{
//	    Name:        "Tutorials",
//	    Slug:        "tutorials",
//	    Description: "Step-by-step guides",
//	})
//
// The package provides a flexible and extensible CMS suitable for blogs, documentation,
// knowledge bases, and other content-driven applications.
package content

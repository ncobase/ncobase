package mixin

import (
	"stocms/pkg/nanoid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserID adds user id field.
type UserID struct{ ent.Schema }

// Fields of the user id mixin.
func (UserID) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").Comment("user id").Optional().MaxLen(nanoid.PrimaryKeySize), // related user id
	}
}

// Indexes of the UserID.
func (UserID) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}

// user id mixin must implement `Mixin` interface.
var _ ent.Mixin = (*UserID)(nil)

// RoleID adds role id field.
type RoleID struct{ ent.Schema }

// Fields of the role id mixin.
func (RoleID) Fields() []ent.Field {
	return []ent.Field{
		field.String("role_id").Comment("role id").Optional().MaxLen(nanoid.PrimaryKeySize), // related role id
	}
}

// Indexes of the RoleID.
func (RoleID) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_id"),
	}
}

// role id mixin must implement `Mixin` interface.
var _ ent.Mixin = (*RoleID)(nil)

// PermissionID adds permission id field.
type PermissionID struct{ ent.Schema }

// Fields of the permission id mixin.
func (PermissionID) Fields() []ent.Field {
	return []ent.Field{
		field.String("permission_id").Comment("permission id").Optional().MaxLen(nanoid.PrimaryKeySize), // related permission id
	}
}

// Indexes of the PermissionID.
func (PermissionID) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("permission_id"),
	}
}

// permission id mixin must implement `Mixin` interface.
var _ ent.Mixin = (*PermissionID)(nil)

// GroupID adds group id field.
type GroupID struct{ ent.Schema }

// Fields of the group id mixin.
func (GroupID) Fields() []ent.Field {
	return []ent.Field{
		field.String("group_id").Comment("group id").Optional().MaxLen(nanoid.PrimaryKeySize), // related group id
	}
}

// Indexes of the GroupID.
func (GroupID) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group_id"),
	}
}

// group id mixin must implement `Mixin` interface.
var _ ent.Mixin = (*GroupID)(nil)

// DomainID adds domain id field.
type DomainID struct{ ent.Schema }

// Fields of the domain id mixin.
func (DomainID) Fields() []ent.Field {
	return []ent.Field{
		field.String("domain_id").Comment("domain id").Optional().MaxLen(nanoid.PrimaryKeySize), // related domain id
	}
}

// Indexes of the DomainID.
func (DomainID) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("domain_id"),
	}
}

// domain id mixin must implement `Mixin` interface.
var _ ent.Mixin = (*DomainID)(nil)

// OrganizationID adds organization id field.
type OrganizationID struct{ ent.Schema }

// Fields of the organization id mixin.
func (OrganizationID) Fields() []ent.Field {
	return []ent.Field{
		field.String("organization_id").Comment("organization id").Optional().MaxLen(nanoid.PrimaryKeySize), // related organization id
	}
}

// Indexes of the OrganizationID.
func (OrganizationID) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("organization_id"),
	}
}

// organization id mixin must implement `Mixin` interface.
var _ ent.Mixin = (*OrganizationID)(nil)

// ParentID adds parent id field.
type ParentID struct{ ent.Schema }

// Fields of the parent id mixin.
func (ParentID) Fields() []ent.Field {
	return []ent.Field{
		field.String("parent_id").Comment("parent id").Optional().MaxLen(nanoid.PrimaryKeySize), // related parent id
	}
}

// Indexes of the ParentID.
func (ParentID) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parent_id"),
	}
}

// parent id mixin must implement `Mixin` interface.
var _ ent.Mixin = (*ParentID)(nil)

// TopicID adds topic id field.
type TopicID struct{ ent.Schema }

// Fields of the topic id mixin.
func (TopicID) Fields() []ent.Field {
	return []ent.Field{
		field.String("topic_id").Comment("topic id").Optional().MaxLen(nanoid.PrimaryKeySize), // related topic id
	}
}

// Indexes of the TopicID.
func (TopicID) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("topic_id"),
	}
}

// topic id mixin must implement `Mixin` interface.
var _ ent.Mixin = (*TopicID)(nil)

// ReplyTo adds reply to field.
type ReplyTo struct{ ent.Schema }

// Fields of the reply to mixin.
func (ReplyTo) Fields() []ent.Field {
	return []ent.Field{
		field.String("reply_to").Comment("reply to object id").Optional().MaxLen(nanoid.PrimaryKeySize), // reply to object id
	}
}

// Indexes of the ReplyTo.
func (ReplyTo) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("reply_to"),
	}
}

// reply to mixin must implement `Mixin` interface.
var _ ent.Mixin = (*ReplyTo)(nil)

// TaxonomyID adds taxonomy id field.
type TaxonomyID struct{ ent.Schema }

// Fields of the taxonomy id mixin.
func (TaxonomyID) Fields() []ent.Field {
	return []ent.Field{
		field.String("taxonomy_id").Comment("taxonomy id").Optional().MaxLen(nanoid.PrimaryKeySize), // related taxonomy id
	}
}

// Indexes of the TaxonomyID.
func (TaxonomyID) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("taxonomy_id"),
	}
}

// taxonomy id mixin must implement `Mixin` interface.
var _ ent.Mixin = (*TaxonomyID)(nil)

// StoreID adds store id field.
type StoreID struct{ ent.Schema }

// Fields of the store id mixin.
func (StoreID) Fields() []ent.Field {
	return []ent.Field{
		field.String("store_id").Comment("store id").Optional().MaxLen(nanoid.PrimaryKeySize), // related store id
	}
}

// Indexes of the StoreID.
func (StoreID) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("store_id"),
	}
}

// store id mixin must implement `Mixin` interface.
var _ ent.Mixin = (*StoreID)(nil)

// CatalogID adds catalog id field.
type CatalogID struct{ ent.Schema }

// Fields of the catalog id mixin.
func (CatalogID) Fields() []ent.Field {
	return []ent.Field{
		field.String("catalog_id").Comment("catalog id").Optional().MaxLen(nanoid.PrimaryKeySize), // related catalog id
	}
}

// Indexes of the CatalogID.
func (CatalogID) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("catalog_id"),
	}
}

// catalog id mixin must implement `Mixin` interface.
var _ ent.Mixin = (*CatalogID)(nil)

// ObjectID adds object id field.
type ObjectID struct{ ent.Schema }

// Fields of the object id mixin.
func (ObjectID) Fields() []ent.Field {
	return []ent.Field{
		field.String("object_id").Comment("object id").Optional().MaxLen(nanoid.PrimaryKeySize), // related object id
	}
}

// Indexes of the ObjectID.
func (ObjectID) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("object_id"),
	}
}

// object id mixin must implement `Mixin` interface.
var _ ent.Mixin = (*ObjectID)(nil)

// OAuthID adds oauth id field.
type OAuthID struct{ ent.Schema }

// Fields of the oauth id mixin.
func (OAuthID) Fields() []ent.Field {
	return []ent.Field{
		field.String("oauth_id").Comment("oauth id").Optional(), // related oauth id
	}
}

// Indexes of the OAuthID.
func (OAuthID) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("oauth_id"),
	}
}

// oauth id mixin must implement `Mixin` interface.
var _ ent.Mixin = (*OAuthID)(nil)

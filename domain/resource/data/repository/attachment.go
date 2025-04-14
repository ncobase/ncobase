package repository

import (
	"context"
	"fmt"
	"ncobase/domain/resource/data"
	"ncobase/domain/resource/data/ent"
	attachmentEnt "ncobase/domain/resource/data/ent/attachment"
	"ncobase/domain/resource/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/data/search/meili"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
)

// AttachmentRepositoryInterface represents the attachment repository interface.
type AttachmentRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateAttachmentBody) (*ent.Attachment, error)
	GetByID(ctx context.Context, slug string) (*ent.Attachment, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Attachment, error)
	Delete(ctx context.Context, slug string) error
	List(ctx context.Context, params *structs.ListAttachmentParams) ([]*ent.Attachment, error)
	CountX(ctx context.Context, params *structs.ListAttachmentParams) int
}

// attachmentRepostory implements the AttachmentRepositoryInterface.
type attachmentRepostory struct {
	ec  *ent.Client
	ecr *ent.Client
	rc  *redis.Client
	ms  *meili.Client
	c   *cache.Cache[ent.Attachment]
}

// NewAttachmentRepository creates a new attachment repository.
func NewAttachmentRepository(d *data.Data) AttachmentRepositoryInterface {
	ec := d.GetEntClient()
	ecr := d.GetEntClientRead()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &attachmentRepostory{ec, ecr, rc, ms, cache.NewCache[ent.Attachment](rc, "ncse_attachment")}
}

// Create creates an attachment.
func (r *attachmentRepostory) Create(ctx context.Context, body *structs.CreateAttachmentBody) (*ent.Attachment, error) {

	// create builder.
	builder := r.ec.Attachment.Create()
	// set values.

	builder.SetNillableName(&body.Name)
	builder.SetNillablePath(&body.Path)
	builder.SetNillableType(&body.Type)
	builder.SetNillableSize(body.Size)
	builder.SetNillableStorage(&body.Storage)
	builder.SetNillableBucket(&body.Bucket)
	builder.SetNillableEndpoint(&body.Endpoint)
	builder.SetNillableObjectID(&body.ObjectID)
	builder.SetNillableTenantID(&body.TenantID)
	builder.SetNillableCreatedBy(body.CreatedBy)

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "attachmentRepo.Create error: %v", err)
		return nil, err
	}

	// create the attachment in Meilisearch index
	if err = r.ms.IndexDocuments("attachments", row); err != nil {
		logger.Errorf(ctx, "attachmentRepo.Create index error: %v", err)
		// return nil, err
	}

	return row, nil
}

// GetByID gets an attachment by ID.
func (r *attachmentRepostory) GetByID(ctx context.Context, slug string) (*ent.Attachment, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", slug)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindAttachment(ctx, &structs.FindAttachment{Attachment: slug})
	if err != nil {
		logger.Errorf(ctx, "attachmentRepo.GetByID error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "attachmentRepo.GetByID cache error: %v", err)
	}

	return row, nil
}

// Update updates an attachment by ID.
func (r *attachmentRepostory) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Attachment, error) {
	attachment, err := r.FindAttachment(ctx, &structs.FindAttachment{Attachment: slug})
	if err != nil {
		return nil, err
	}

	// create builder.
	builder := r.ec.Attachment.UpdateOne(attachment)

	// set values
	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(types.ToPointer(value.(string)))
		case "path":
			builder.SetNillablePath(types.ToPointer(value.(string)))
		case "type":
			builder.SetNillableType(types.ToPointer(value.(string)))
		case "size":
			builder.SetNillableSize(types.ToPointer(value.(int)))
		case "storage":
			builder.SetNillableStorage(types.ToPointer(value.(string)))
		case "endpoint":
			builder.SetNillableEndpoint(types.ToPointer(value.(string)))
		case "object_id":
			builder.SetNillableObjectID(types.ToPointer(value.(string)))
		case "tenant_id":
			builder.SetNillableTenantID(types.ToPointer(value.(string)))
		case "extras":
			builder.SetExtras(value.(types.JSON))
		case "updated_by":
			builder.SetNillableUpdatedBy(types.ToPointer(value.(string)))
		}
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "attachmentRepo.Update error: %v", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", attachment.ID)
	if err = r.c.Delete(ctx, cacheKey); err != nil {
		logger.Errorf(ctx, "attachmentRepo.Update cache error: %v", err)
	}

	// delete from Meilisearch index
	if err = r.ms.DeleteDocuments("attachments", attachment.ID); err != nil {
		logger.Errorf(ctx, "attachmentRepo.Update index error: %v", err)
		// return nil, err
	}

	return row, nil
}

// Delete deletes an attachment by ID.
func (r *attachmentRepostory) Delete(ctx context.Context, slug string) error {
	attachment, err := r.FindAttachment(ctx, &structs.FindAttachment{Attachment: slug})
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.Attachment.Delete()

	// execute the builder and verify the result.
	if _, err = builder.Where(attachmentEnt.IDEQ(slug)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "attachmentRepo.Delete error: %v", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", attachment.ID)
	if err = r.c.Delete(ctx, cacheKey); err != nil {
		logger.Errorf(ctx, "attachmentRepo.Delete cache error: %v", err)
	}

	// delete from Meilisearch index
	if err = r.ms.DeleteDocuments("attachments", attachment.ID); err != nil {
		logger.Errorf(ctx, "attachmentRepo.Delete index error: %v", err)
		// return nil, err
	}

	return nil
}

// FindAttachment finds an attachment.
func (r *attachmentRepostory) FindAttachment(ctx context.Context, params *structs.FindAttachment) (*ent.Attachment, error) {
	// create builder.
	builder := r.ecr.Attachment.Query()

	if validator.IsNotEmpty(params.Attachment) {
		builder = builder.Where(attachmentEnt.Or(
			attachmentEnt.IDEQ(params.Attachment),
			attachmentEnt.NameEQ(params.Attachment),
		))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// List gets a list of attachments.
func (r *attachmentRepostory) List(ctx context.Context, params *structs.ListAttachmentParams) ([]*ent.Attachment, error) {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	if params.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, fmt.Errorf("invalid id in cursor: %s", id)
		}

		if params.Direction == "backward" {
			builder.Where(
				attachmentEnt.Or(
					attachmentEnt.CreatedAtGT(timestamp),
					attachmentEnt.And(
						attachmentEnt.CreatedAtEQ(timestamp),
						attachmentEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				attachmentEnt.Or(
					attachmentEnt.CreatedAtLT(timestamp),
					attachmentEnt.And(
						attachmentEnt.CreatedAtEQ(timestamp),
						attachmentEnt.IDLT(id),
					),
				),
			)
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(attachmentEnt.FieldCreatedAt), ent.Asc(attachmentEnt.FieldID))
	} else {
		builder.Order(ent.Desc(attachmentEnt.FieldCreatedAt), ent.Desc(attachmentEnt.FieldID))
	}

	builder.Limit(params.Limit)

	// execute the builder.
	rows, err := builder.All(ctx)
	if validator.IsNotNil(err) {
		logger.Errorf(ctx, "attachmentRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// ListBuilder creates list builder.
func (r *attachmentRepostory) ListBuilder(ctx context.Context, params *structs.ListAttachmentParams) (*ent.AttachmentQuery, error) {
	// create builder.
	builder := r.ecr.Attachment.Query()

	// belong tenant
	if params.Tenant != "" {
		builder = builder.Where(attachmentEnt.TenantIDEQ(params.Tenant))
	}

	// belong user
	if params.User != "" {
		builder = builder.Where(attachmentEnt.CreatedByEQ(params.User))
	}

	// object id
	if params.Object != "" {
		builder = builder.Where(attachmentEnt.ObjectIDEQ(params.Object))
	}

	// attachment type
	if params.Type != "" {
		builder = builder.Where(attachmentEnt.TypeContains(params.Type))
	}

	// storage provider
	if params.Storage != "" {
		builder = builder.Where(attachmentEnt.StorageEQ(params.Storage))
	}

	return builder, nil
}

// CountX counts attachments based on given parameters.
func (r *attachmentRepostory) CountX(ctx context.Context, params *structs.ListAttachmentParams) int {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

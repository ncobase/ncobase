package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"ncobase/common/ecode"
	"ncobase/common/helper"
	"ncobase/common/log"
	"ncobase/common/paging"
	"ncobase/common/validator"
	"ncobase/feature/resource/data"
	"ncobase/feature/resource/data/ent"
	"ncobase/feature/resource/data/repository"
	"ncobase/feature/resource/structs"
	"os"
)

// AttachmentServiceInterface represents the attachment service interface.
type AttachmentServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateAttachmentBody) (*structs.ReadAttachment, error)
	Update(ctx context.Context, slug string, updates map[string]any) (*structs.ReadAttachment, error)
	Get(ctx context.Context, slug string) (*structs.ReadAttachment, error)
	Delete(ctx context.Context, slug string) error
	List(ctx context.Context, params *structs.ListAttachmentParams) (paging.Result[*structs.ReadAttachment], error)
	GetFileStream(ctx context.Context, slug string) (io.ReadCloser, *structs.ReadAttachment, error)
}

// AttachmentService is the struct for the attachment service.
type attachmentService struct {
	attachment repository.AttachmentRepositoryInterface
}

// NewAttachmentService creates a new attachment service.
func NewAttachmentService(d *data.Data) AttachmentServiceInterface {
	return &attachmentService{
		attachment: repository.NewAttachmentRepository(d),
	}
}

// Create creates a new attachment.
func (s *attachmentService) Create(ctx context.Context, body *structs.CreateAttachmentBody) (*structs.ReadAttachment, error) {
	if validator.IsEmpty(body.ObjectID) {
		return nil, errors.New(ecode.FieldIsRequired("belongsTo object"))
	}

	if validator.IsEmpty(body.TenantID) {
		return nil, errors.New(ecode.FieldIsRequired("belongsTo tenant"))
	}
	// get storage interface
	storage, storageConfig := helper.GetStorage(ctx)

	// Handle file storage
	_, err := storage.Put(body.Path, body.File)
	if err != nil {
		log.Errorf(ctx, "Error storing file: %v", err)
		return nil, errors.New("failed to store file")
	}
	defer func() {
		if err != nil {
			_ = storage.Delete(body.Path)
		}
	}()

	// set storage provider
	body.Storage = storageConfig.Provider
	// set bucket
	body.Bucket = storageConfig.Bucket
	// set endpoint
	body.Endpoint = storageConfig.Endpoint
	// set created by
	userID := helper.GetUserID(ctx)
	body.CreatedBy = &userID

	// Create the attachment using the repository
	row, err := s.attachment.Create(ctx, body)
	if err != nil {
		log.Errorf(ctx, "Error creating attachment: %v", err)
		return nil, errors.New("failed to create attachment")
	}

	return s.Serialize(row), nil
}

// Update updates an existing attachment.
func (s *attachmentService) Update(ctx context.Context, slug string, updates map[string]any) (*structs.ReadAttachment, error) {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return nil, errors.New(ecode.FieldIsRequired("slug"))
	}

	// Check if updates map is empty
	if len(updates) == 0 {
		return nil, errors.New(ecode.FieldIsEmpty("updates fields"))
	}

	// Get storage interface
	storage, storageConfig := helper.GetStorage(ctx)

	// Handle file update if path is included in updates
	if path, ok := updates["path"].(string); ok {
		// Check if the file content is included in the updates
		if file, ok := updates["file"].(io.Reader); ok {
			if _, err := storage.Put(path, file); err != nil {
				log.Errorf(ctx, "Error updating file: %v", err)
				return nil, errors.New("error updating file")
			}
			// update storage
			updates["storage"] = storageConfig.Provider
			// update bucket
			updates["bucket"] = storageConfig.Bucket
			// update endpoint
			updates["endpoint"] = storageConfig.Endpoint
			// Remove file from updates after storing to avoid saving the file object itself in DB
			delete(updates, "file")
			// set updated by
			if _, ok := updates["updated_by"].(string); !ok {
				updates["updated_by"] = helper.GetUserID(ctx)
			}
		} else {
			log.Warnf(ctx, "File content is missing, skipping file update")
		}
	}

	row, err := s.attachment.Update(ctx, slug, updates)
	if err := handleEntError(ctx, "Attachment", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves an attachment by ID.
func (s *attachmentService) Get(ctx context.Context, slug string) (*structs.ReadAttachment, error) {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return nil, errors.New(ecode.FieldIsRequired("slug"))
	}

	// get storage interface
	storage, _ := helper.GetStorage(ctx)

	row, err := s.attachment.GetByID(ctx, slug)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New(ecode.NotExist(fmt.Sprintf("Attachment %s", slug)))
		}
		log.Errorf(ctx, "Error retrieving attachment: %v", err)
		return nil, errors.New("error retrieving attachment")
	}

	// Fetch file from storage
	file, err := storage.Get(row.Path)
	if err != nil {
		log.Errorf(ctx, "Error retrieving file: %v", err)
		return nil, errors.New("error retrieving file")
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Errorf(ctx, "Error closing file: %v", err)
		}
	}(file)

	return s.Serialize(row), nil
}

// Delete deletes an attachment by ID.
func (s *attachmentService) Delete(ctx context.Context, slug string) error {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return errors.New(ecode.FieldIsRequired("slug"))
	}

	// get storage interface
	storage, _ := helper.GetStorage(ctx)

	row, err := s.attachment.GetByID(ctx, slug)
	if err != nil {
		log.Errorf(ctx, "Error retrieving attachment: %v", err)
		return errors.New("error retrieving attachment")
	}

	err = s.attachment.Delete(ctx, slug)
	if err != nil {
		log.Errorf(ctx, "Error deleting attachment: %v", err)
		return errors.New("error deleting attachment")
	}

	// Delete the file from storage
	if err := storage.Delete(row.Path); err != nil {
		log.Errorf(ctx, "Error deleting file: %v", err)
		return errors.New("error deleting file")
	}

	return nil
}

// List lists attachments.
func (s *attachmentService) List(ctx context.Context, params *structs.ListAttachmentParams) (paging.Result[*structs.ReadAttachment], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadAttachment, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.attachment.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			log.Errorf(ctx, "Error listing attachments: %v", err)
			return nil, 0, err
		}

		total := s.attachment.CountX(ctx, params)

		return s.Serializes(rows), total, nil
	})
}

// GetFileStream retrieves an attachment's file stream.
func (s *attachmentService) GetFileStream(ctx context.Context, slug string) (io.ReadCloser, *structs.ReadAttachment, error) {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return nil, nil, errors.New(ecode.FieldIsRequired("slug"))
	}

	// Get storage interface
	storage, _ := helper.GetStorage(ctx)

	// Retrieve attachment by ID
	row, err := s.attachment.GetByID(ctx, slug)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil, errors.New(ecode.NotExist(fmt.Sprintf("Attachment %s", slug)))
		}
		log.Errorf(ctx, "Error retrieving attachment: %v", err)
		return nil, nil, errors.New("error retrieving attachment")
	}

	// Fetch file stream from storage
	fileStream, err := storage.GetStream(row.Path)
	if err != nil {
		log.Errorf(ctx, "Error retrieving file stream: %v", err)
		return nil, nil, errors.New("error retrieving file stream")
	}

	// Return file stream along with attachment information
	return fileStream, s.Serialize(row), nil
}

// Serializes serializes attachments.
func (s *attachmentService) Serializes(rows []*ent.Attachment) []*structs.ReadAttachment {
	var rs []*structs.ReadAttachment
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes a attachment.
func (s *attachmentService) Serialize(row *ent.Attachment) *structs.ReadAttachment {
	return &structs.ReadAttachment{
		ID:        row.ID,
		Name:      row.Name,
		Path:      row.Path,
		Type:      row.Type,
		Size:      &row.Size,
		Storage:   row.Storage,
		Bucket:    row.Bucket,
		Endpoint:  row.Endpoint,
		ObjectID:  row.ObjectID,
		TenantID:  row.TenantID,
		Extras:    &row.Extras,
		CreatedBy: &row.CreatedBy,
		CreatedAt: &row.CreatedAt,
		UpdatedBy: &row.UpdatedBy,
		UpdatedAt: &row.UpdatedAt,
	}
}

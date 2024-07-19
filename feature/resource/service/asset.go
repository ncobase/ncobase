package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/paging"
	"ncobase/common/validator"
	"ncobase/feature/resource/data"
	"ncobase/feature/resource/data/ent"
	"ncobase/feature/resource/data/repository"
	"ncobase/feature/resource/structs"
	"ncobase/helper"
	"os"
)

// AssetServiceInterface represents the asset service interface.
type AssetServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateAssetBody) (*structs.ReadAsset, error)
	Update(ctx context.Context, slug string, updates map[string]any) (*structs.ReadAsset, error)
	Get(ctx context.Context, slug string) (*structs.ReadAsset, error)
	Delete(ctx context.Context, slug string) error
	List(ctx context.Context, params *structs.ListAssetParams) (paging.Result[*structs.ReadAsset], error)
	GetFileStream(ctx context.Context, slug string) (io.ReadCloser, *structs.ReadAsset, error)
}

// AssetService is the struct for the asset service.
type assetService struct {
	asset repository.AssetRepositoryInterface
}

// NewAssetService creates a new asset service.
func NewAssetService(d *data.Data) AssetServiceInterface {
	return &assetService{
		asset: repository.NewAssetRepository(d),
	}
}

// Create creates a new asset.
func (s *assetService) Create(ctx context.Context, body *structs.CreateAssetBody) (*structs.ReadAsset, error) {
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
		log.Errorf(ctx, "Error storing file: %v\n", err)
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

	// Create the asset using the repository
	row, err := s.asset.Create(ctx, body)
	if err != nil {
		log.Errorf(ctx, "Error creating asset: %v\n", err)
		return nil, errors.New("failed to create asset")
	}

	return s.Serialize(row), nil
}

// Update updates an existing asset.
func (s *assetService) Update(ctx context.Context, slug string, updates map[string]any) (*structs.ReadAsset, error) {
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
				log.Errorf(ctx, "Error updating file: %v\n", err)
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

	row, err := s.asset.Update(ctx, slug, updates)
	if err := handleEntError("Asset", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves an asset by ID.
func (s *assetService) Get(ctx context.Context, slug string) (*structs.ReadAsset, error) {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return nil, errors.New(ecode.FieldIsRequired("slug"))
	}

	// get storage interface
	storage, _ := helper.GetStorage(ctx)

	row, err := s.asset.GetByID(ctx, slug)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New(ecode.NotExist(fmt.Sprintf("Asset %s", slug)))
		}
		log.Errorf(ctx, "Error retrieving asset: %v\n", err)
		return nil, errors.New("error retrieving asset")
	}

	// Fetch file from storage
	file, err := storage.Get(row.Path)
	if err != nil {
		log.Errorf(ctx, "Error retrieving file: %v\n", err)
		return nil, errors.New("error retrieving file")
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Errorf(ctx, "Error closing file: %v\n", err)
		}
	}(file)

	return s.Serialize(row), nil
}

// Delete deletes an asset by ID.
func (s *assetService) Delete(ctx context.Context, slug string) error {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return errors.New(ecode.FieldIsRequired("slug"))
	}

	// get storage interface
	storage, _ := helper.GetStorage(ctx)

	row, err := s.asset.GetByID(ctx, slug)
	if err != nil {
		log.Errorf(ctx, "Error retrieving asset: %v\n", err)
		return errors.New("error retrieving asset")
	}

	err = s.asset.Delete(ctx, slug)
	if err != nil {
		log.Errorf(ctx, "Error deleting asset: %v\n", err)
		return errors.New("error deleting asset")
	}

	// Delete the file from storage
	if err := storage.Delete(row.Path); err != nil {
		log.Errorf(ctx, "Error deleting file: %v\n", err)
		return errors.New("error deleting file")
	}

	return nil
}

// List lists assets.
func (s *assetService) List(ctx context.Context, params *structs.ListAssetParams) (paging.Result[*structs.ReadAsset], error) {
	pp := paging.Params{
		Cursor: params.Cursor,
		Limit:  params.Limit,
	}

	return paging.Paginate(pp, func(cursor string, offset int, limit int, direction string) ([]*structs.ReadAsset, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Offset = offset
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.asset.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			log.Errorf(ctx, "Error listing assets: %v\n", err)
			return nil, 0, err
		}

		total := s.asset.CountX(ctx, params)

		return s.Serializes(rows), total, nil
	})
}

// GetFileStream retrieves an asset's file stream.
func (s *assetService) GetFileStream(ctx context.Context, slug string) (io.ReadCloser, *structs.ReadAsset, error) {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return nil, nil, errors.New(ecode.FieldIsRequired("slug"))
	}

	// Get storage interface
	storage, _ := helper.GetStorage(ctx)

	// Retrieve asset by ID
	row, err := s.asset.GetByID(ctx, slug)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil, errors.New(ecode.NotExist(fmt.Sprintf("Asset %s", slug)))
		}
		log.Errorf(ctx, "Error retrieving asset: %v\n", err)
		return nil, nil, errors.New("error retrieving asset")
	}

	// Fetch file stream from storage
	fileStream, err := storage.GetStream(row.Path)
	if err != nil {
		log.Errorf(ctx, "Error retrieving file stream: %v\n", err)
		return nil, nil, errors.New("error retrieving file stream")
	}

	// Return file stream along with asset information
	return fileStream, s.Serialize(row), nil
}

// Serializes serializes assets.
func (s *assetService) Serializes(rows []*ent.Asset) []*structs.ReadAsset {
	var rs []*structs.ReadAsset
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes a asset.
func (s *assetService) Serialize(row *ent.Asset) *structs.ReadAsset {
	return &structs.ReadAsset{
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

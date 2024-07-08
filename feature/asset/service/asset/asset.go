package asset

import (
	"context"
	"fmt"
	"io"
	"ncobase/feature/asset/data"
	"ncobase/feature/asset/data/ent"
	"ncobase/feature/asset/data/repository/asset"
	"ncobase/feature/asset/structs"
	"ncobase/helper"
	"os"

	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/common/validator"
)

// ServiceInterface represents the asset service interface.
type ServiceInterface interface {
	CreateAssetService(ctx context.Context, body *structs.CreateAssetBody) (*resp.Exception, error)
	UpdateAssetService(ctx context.Context, slug string, updates map[string]any) (*resp.Exception, error)
	GetAssetService(ctx context.Context, slug string) (*resp.Exception, error)
	DeleteAssetService(ctx context.Context, slug string) (*resp.Exception, error)
	ListAssetsService(ctx context.Context, params *structs.ListAssetParams) (*resp.Exception, error)
	GetFileStream(ctx context.Context, slug string) (io.ReadCloser, *resp.Exception)
}

type Service struct {
	asset asset.RepositoryInterface
}

func New(d *data.Data) ServiceInterface {
	return &Service{
		asset: asset.NewAsset(d),
	}
}

// CreateAssetService creates a new asset.
func (svc *Service) CreateAssetService(ctx context.Context, body *structs.CreateAssetBody) (*resp.Exception, error) {
	if validator.IsEmpty(body.ObjectID) {
		return resp.BadRequest(ecode.FieldIsRequired("belongsTo object")), nil
	}

	if validator.IsEmpty(body.TenantID) {
		return resp.BadRequest(ecode.FieldIsRequired("belongsTo tenant")), nil
	}
	// get storage interface
	storage, storageConfig := helper.GetStorage(ctx)

	// Handle file storage
	_, err := storage.Put(body.Path, body.File)
	if err != nil {
		log.Errorf(ctx, "Error storing file: %v\n", err)
		return resp.InternalServer("failed to store file"), nil
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
	row, err := svc.asset.Create(ctx, body)
	if err != nil {
		log.Errorf(ctx, "Error creating asset: %v\n", err)
		return resp.InternalServer("failed to create asset"), nil
	}

	return &resp.Exception{
		Data: row,
	}, nil
}

// UpdateAssetService updates an existing asset.
func (svc *Service) UpdateAssetService(ctx context.Context, slug string, updates map[string]any) (*resp.Exception, error) {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return resp.BadRequest(ecode.FieldIsRequired("slug")), nil
	}

	// Check if updates map is empty
	if len(updates) == 0 {
		return resp.BadRequest(ecode.FieldIsEmpty("updates fields")), nil
	}

	// Get storage interface
	storage, storageConfig := helper.GetStorage(ctx)

	// Handle file update if path is included in updates
	if path, ok := updates["path"].(string); ok {
		// Check if the file content is included in the updates
		if file, ok := updates["file"].(io.Reader); ok {
			if _, err := storage.Put(path, file); err != nil {
				log.Errorf(ctx, "Error updating file: %v\n", err)
				return resp.InternalServer("Error updating file"), err
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

	row, err := svc.asset.Update(ctx, slug, updates)
	if exception, err := helper.HandleError("Asset", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: row,
	}, nil
}

// GetAssetService retrieves an asset by ID.
func (svc *Service) GetAssetService(ctx context.Context, slug string) (*resp.Exception, error) {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return resp.BadRequest(ecode.FieldIsRequired("slug")), nil
	}

	// get storage interface
	storage, _ := helper.GetStorage(ctx)

	row, err := svc.asset.GetByID(ctx, slug)
	if err != nil {
		if ent.IsNotFound(err) {
			return resp.NotFound(ecode.NotExist(fmt.Sprintf("Asset %s", slug))), nil
		}
		log.Errorf(ctx, "Error retrieving asset: %v\n", err)
		return resp.InternalServer("Error retrieving asset"), err
	}

	// Fetch file from storage
	file, err := storage.Get(row.Path)
	if err != nil {
		log.Errorf(ctx, "Error retrieving file: %v\n", err)
		return resp.InternalServer("Error retrieving file"), err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Errorf(ctx, "Error closing file: %v\n", err)
		}
	}(file)

	return &resp.Exception{
		Data: row,
	}, nil
}

// DeleteAssetService deletes an asset by ID.
func (svc *Service) DeleteAssetService(ctx context.Context, slug string) (*resp.Exception, error) {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return resp.BadRequest(ecode.FieldIsRequired("slug")), nil
	}

	// get storage interface
	storage, _ := helper.GetStorage(ctx)

	row, err := svc.asset.GetByID(ctx, slug)
	if err != nil {
		log.Errorf(ctx, "Error retrieving asset: %v\n", err)
		return resp.InternalServer("Error retrieving asset"), err
	}

	err = svc.asset.Delete(ctx, slug)
	if err != nil {
		log.Errorf(ctx, "Error deleting asset: %v\n", err)
		return resp.InternalServer("Error deleting asset"), err
	}

	// Delete the file from storage
	if err := storage.Delete(row.Path); err != nil {
		log.Errorf(ctx, "Error deleting file: %v\n", err)
		return resp.InternalServer("Error deleting file"), err
	}

	return nil, nil
}

// ListAssetsService lists assets.
func (svc *Service) ListAssetsService(ctx context.Context, params *structs.ListAssetParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	rows, err := svc.asset.List(ctx, params)

	if ent.IsNotFound(err) {
		return resp.NotFound(ecode.FieldIsInvalid("cursor")), nil
	}
	if validator.IsNotNil(err) {
		log.Errorf(ctx, "Error listing assets: %v\n", err)
		return resp.InternalServer(err.Error()), nil
	}

	total := svc.asset.CountX(ctx, params)

	return &resp.Exception{
		Data: types.JSON{
			"content": rows,
			"total":   total,
		},
	}, nil
}

// GetFileStream retrieves an asset's file stream.
func (svc *Service) GetFileStream(ctx context.Context, slug string) (io.ReadCloser, *resp.Exception) {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return nil, resp.BadRequest(ecode.FieldIsRequired("slug"))
	}

	// Get storage interface
	storage, _ := helper.GetStorage(ctx)

	// Retrieve asset by ID
	row, err := svc.asset.GetByID(ctx, slug)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, resp.NotFound(ecode.NotExist(fmt.Sprintf("Asset %s", slug)))
		}
		log.Errorf(ctx, "Error retrieving asset: %v\n", err)
		return nil, resp.InternalServer("Error retrieving asset")
	}

	// Fetch file stream from storage
	fileStream, err := storage.GetStream(row.Path)
	if err != nil {
		log.Errorf(ctx, "Error retrieving file stream: %v\n", err)
		return nil, resp.InternalServer("Error retrieving file stream")
	}

	// Return file stream along with asset information
	return fileStream, &resp.Exception{
		Data: row,
	}
}

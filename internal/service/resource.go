package service

import (
	"fmt"
	"io"
	"os"
	"stocms/internal/data/ent"
	"stocms/internal/data/structs"
	"stocms/internal/helper"
	"stocms/pkg/ecode"
	"stocms/pkg/log"
	"stocms/pkg/resp"
	"stocms/pkg/validator"

	"github.com/gin-gonic/gin"
)

// CreateResourceService creates a new resource.
func (svc *Service) CreateResourceService(c *gin.Context, body *structs.CreateResourceBody) (*resp.Exception, error) {
	// get storage interface
	storage, storageConfig := helper.GetStorage(c)

	// Handle file storage
	obj, err := storage.Put(body.Path, body.File)
	if err != nil {
		log.Errorf(c, "Error storing file: %v\n", err)
		return resp.InternalServer("Error storing file"), err
	}

	fmt.Printf("obj: %v\n", obj)

	// set storage provider
	body.Storage = storageConfig.Provider
	// set created by
	body.CreatedBy = helper.GetUserID(c)

	// Create the resource using the repository
	resource, err := svc.resource.Create(c, body)
	if err != nil {
		log.Errorf(c, "Error creating resource: %v\n", err)
		// delete file from storage
		_ = storage.Delete(body.Path)
		return resp.InternalServer("Error creating resource"), err
	}

	return &resp.Exception{
		Data: resource,
	}, nil
}

// UpdateResourceService updates an existing resource.
func (svc *Service) UpdateResourceService(c *gin.Context, slug string, updates map[string]interface{}) (*resp.Exception, error) {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return resp.BadRequest(ecode.FieldIsRequired("slug")), nil
	}

	// Check if updates map is empty
	if len(updates) == 0 {
		return resp.BadRequest(ecode.FieldIsRequired("updates")), nil
	}

	// Get storage interface
	storage, _ := helper.GetStorage(c)

	// Handle file update if path is included in updates
	if path, ok := updates["path"].(string); ok {
		// Check if the file content is included in the updates
		if file, ok := updates["file"].(io.Reader); ok {
			if _, err := storage.Put(path, file); err != nil {
				log.Errorf(c, "Error updating file: %v\n", err)
				return resp.InternalServer("Error updating file"), err
			}
			// Remove file from updates after storing to avoid saving the file object itself in DB
			delete(updates, "file")
		} else {
			log.Warnf(c, "File content is missing, skipping file update")
		}
	}

	// Call the repository's Update method
	resource, err := svc.resource.Update(c, slug, updates)
	if exception, err := handleError("Resource", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: resource,
	}, nil
}

// GetResourceService retrieves an resource by ID.
func (svc *Service) GetResourceService(c *gin.Context, slug string) (*resp.Exception, error) {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return resp.BadRequest(ecode.FieldIsRequired("slug")), nil
	}

	// get storage interface
	storage, _ := helper.GetStorage(c)

	// Call the repository'storage GetByID method
	resource, err := svc.resource.GetByID(c, slug)
	if err != nil {
		if ent.IsNotFound(err) {
			return resp.NotFound(ecode.NotExist(fmt.Sprintf("Resource %s", slug))), nil
		}
		log.Errorf(c, "Error retrieving resource: %v\n", err)
		return resp.InternalServer("Error retrieving resource"), err
	}

	// Fetch file from storage
	file, err := storage.Get(resource.Path)
	if err != nil {
		log.Errorf(c, "Error retrieving file: %v\n", err)
		return resp.InternalServer("Error retrieving file"), err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Errorf(c, "Error closing file: %v\n", err)
		}
	}(file)

	return &resp.Exception{
		Data: resource,
	}, nil
}

// DeleteResourceService deletes an resource by ID.
func (svc *Service) DeleteResourceService(c *gin.Context, slug string) (*resp.Exception, error) {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return resp.BadRequest(ecode.FieldIsRequired("slug")), nil
	}

	// get storage interface
	storage, _ := helper.GetStorage(c)

	// Call the repository'storage GetByID method to get the resource details
	resource, err := svc.resource.GetByID(c, slug)
	if err != nil {
		log.Errorf(c, "Error retrieving resource: %v\n", err)
		return resp.InternalServer("Error retrieving resource"), err
	}

	// Call the repository'storage Delete method
	err = svc.resource.Delete(c, slug)
	if err != nil {
		log.Errorf(c, "Error deleting resource: %v\n", err)
		return resp.InternalServer("Error deleting resource"), err
	}

	// Delete the file from storage
	if err := storage.Delete(resource.Path); err != nil {
		log.Errorf(c, "Error deleting file: %v\n", err)
		return resp.InternalServer("Error deleting file"), err
	}

	return nil, nil
}

// ListResourcesService lists resources.
func (svc *Service) ListResourcesService(c *gin.Context, params *structs.ListResourceParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}
	rows, err := svc.resource.List(c, params)

	if ent.IsNotFound(err) {
		return resp.NotFound(ecode.FieldIsInvalid("cursor")), nil
	}
	if validator.IsNotNil(err) {
		log.Errorf(c, "Error listing resources: %v\n", err)
		return resp.InternalServer(err.Error()), nil
	}

	return &resp.Exception{Data: rows}, nil
}

// GetFileStream retrieves an resource's file stream.
func (svc *Service) GetFileStream(c *gin.Context, slug string) (io.ReadCloser, *resp.Exception) {
	// Check if ID is empty
	if validator.IsEmpty(slug) {
		return nil, resp.BadRequest(ecode.FieldIsRequired("slug"))
	}

	// Get storage interface
	storage, _ := helper.GetStorage(c)

	// Retrieve resource by ID
	resource, err := svc.resource.GetByID(c, slug)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, resp.NotFound(ecode.NotExist(fmt.Sprintf("Resource %s", slug)))
		}
		log.Errorf(c, "Error retrieving resource: %v\n", err)
		return nil, resp.InternalServer("Error retrieving resource")
	}

	// Fetch file stream from storage
	fileStream, err := storage.GetStream(resource.Path)
	if err != nil {
		log.Errorf(c, "Error retrieving file stream: %v\n", err)
		return nil, resp.InternalServer("Error retrieving file stream")
	}

	// Return file stream along with resource information
	return fileStream, &resp.Exception{
		Data: resource,
	}
}

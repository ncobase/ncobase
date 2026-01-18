package handler

import (
	"ncobase/core/user/service"
	"ncobase/core/user/structs"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// ApiKeyHandlerInterface defines handler operations for API keys
type ApiKeyHandlerInterface interface {
	GenerateApiKey(c *gin.Context)
	GetApiKey(c *gin.Context)
	GetUserApiKeys(c *gin.Context)
	GetMyApiKeys(c *gin.Context)
	DeleteApiKey(c *gin.Context)
}

// apiKeyHandler implements ApiKeyHandlerInterface
type apiKeyHandler struct {
	s *service.Service
}

// NewApiKeyHandler creates a new API key handler
func NewApiKeyHandler(svc *service.Service) ApiKeyHandlerInterface {
	return &apiKeyHandler{
		s: svc,
	}
}

// GenerateApiKey generates a new API key
//
// @Summary Generate API key
// @Description Generate a new API key for the current user
// @Tags sys
// @Accept json
// @Produce json
// @Param request body structs.CreateApiKeyRequest true "API key request"
// @Success 200 {object} structs.ApiKey "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/api-keys [post]
// @Security Bearer
func (h *apiKeyHandler) GenerateApiKey(c *gin.Context) {
	var request structs.CreateApiKeyRequest
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, &request); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	// Get user ID from context
	userID := ctxutil.GetUserID(c.Request.Context())
	if userID == "" {
		resp.Fail(c.Writer, resp.UnAuthorized("User not authenticated"))
		return
	}

	result, err := h.s.ApiKey.GenerateApiKey(c.Request.Context(), userID, &request)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetApiKey retrieves an API key by ID
//
// @Summary Get API key
// @Description Retrieve an API key by its ID
// @Tags sys
// @Produce json
// @Param id path string true "API key ID"
// @Success 200 {object} structs.ApiKey "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/api-keys/{id} [get]
// @Security Bearer
func (h *apiKeyHandler) GetApiKey(c *gin.Context) {
	result, err := h.s.ApiKey.GetApiKey(c.Request.Context(), c.Param("id"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetUserApiKeys retrieves all API keys for a specific user
//
// @Summary Get user API keys
// @Description Retrieve all API keys for a specific user
// @Tags sys
// @Produce json
// @Param username path string true "Username"
// @Success 200 {array} structs.ApiKey "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/{username}/api-keys [get]
// @Security Bearer
func (h *apiKeyHandler) GetUserApiKeys(c *gin.Context) {
	// Get user by username
	username := c.Param("username")
	user, err := h.s.User.Get(c.Request.Context(), username)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.s.ApiKey.GetUserApiKeys(c.Request.Context(), user.ID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetMyApiKeys retrieves all API keys for the current user
//
// @Summary Get my API keys
// @Description Retrieve all API keys for the current user
// @Tags sys
// @Produce json
// @Success 200 {array} structs.ApiKey "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/me/api-keys [get]
// @Security Bearer
func (h *apiKeyHandler) GetMyApiKeys(c *gin.Context) {
	// Get user ID from context
	userID := ctxutil.GetUserID(c.Request.Context())
	if userID == "" {
		resp.Fail(c.Writer, resp.UnAuthorized("User not authenticated"))
		return
	}

	result, err := h.s.ApiKey.GetUserApiKeys(c.Request.Context(), userID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// DeleteApiKey removes an API key
//
// @Summary Delete API key
// @Description Delete an API key by its ID
// @Tags sys
// @Produce json
// @Param id path string true "API key ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/api-keys/{id} [delete]
// @Security Bearer
func (h *apiKeyHandler) DeleteApiKey(c *gin.Context) {
	err := h.s.ApiKey.DeleteApiKey(c.Request.Context(), c.Param("id"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

package handler

import (
	"ncobase/core/space/service"
	"ncobase/core/space/structs"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// SpaceQuotaHandlerInterface defines the interface for space quota handler
type SpaceQuotaHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	UpdateUsage(c *gin.Context)
	CheckLimit(c *gin.Context)
	GetSummary(c *gin.Context)
}

// spaceQuotaHandler implements SpaceQuotaHandlerInterface
type spaceQuotaHandler struct {
	s *service.Service
}

// NewSpaceQuotaHandler creates a new space quota handler
func NewSpaceQuotaHandler(svc *service.Service) SpaceQuotaHandlerInterface {
	return &spaceQuotaHandler{s: svc}
}

// Create handles creating a space quota
//
// @Summary Create space quota
// @Description Create a new space quota configuration
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.CreateSpaceQuotaBody true "Quota configuration"
// @Success 200 {object} structs.ReadSpaceQuota "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/quotas [post]
// @Security Bearer
func (h *spaceQuotaHandler) Create(c *gin.Context) {
	body := &structs.CreateSpaceQuotaBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.SpaceQuota.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating a space quota
//
// @Summary Update space quota
// @Description Update an existing space quota configuration
// @Tags sys
// @Accept json
// @Produce json
// @Param id path string true "Quota ID"
// @Param body body types.JSON true "Update data"
// @Success 200 {object} structs.ReadSpaceQuota "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/quotas/{id} [put]
// @Security Bearer
func (h *spaceQuotaHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	updates := &types.JSON{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, updates); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.SpaceQuota.Update(c.Request.Context(), id, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles retrieving a space quota
//
// @Summary Get space quota
// @Description Retrieve a space quota by ID
// @Tags sys
// @Produce json
// @Param id path string true "Quota ID"
// @Success 200 {object} structs.ReadSpaceQuota "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/quotas/{id} [get]
// @Security Bearer
func (h *spaceQuotaHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.s.SpaceQuota.Get(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a space quota
//
// @Summary Delete space quota
// @Description Delete a space quota configuration
// @Tags sys
// @Produce json
// @Param id path string true "Quota ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/quotas/{id} [delete]
// @Security Bearer
func (h *spaceQuotaHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.s.SpaceQuota.Delete(c.Request.Context(), id); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing space quotas
//
// @Summary List space quotas
// @Description Retrieve a list of space quotas
// @Tags sys
// @Produce json
// @Param params query structs.ListSpaceQuotaParams true "List parameters"
// @Success 200 {array} structs.ReadSpaceQuota "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/quotas [get]
// @Security Bearer
func (h *spaceQuotaHandler) List(c *gin.Context) {
	params := &structs.ListSpaceQuotaParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.SpaceQuota.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// UpdateUsage handles updating quota usage
//
// @Summary Update quota usage
// @Description Update the current usage of a quota
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.QuotaUsageRequest true "Usage update request"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/quotas/usage [post]
// @Security Bearer
func (h *spaceQuotaHandler) UpdateUsage(c *gin.Context) {
	body := &structs.QuotaUsageRequest{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	if err := h.s.SpaceQuota.UpdateUsage(c.Request.Context(), body.SpaceID, string(body.QuotaType), body.Delta); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// CheckLimit handles checking quota limits
//
// @Summary Check quota limit
// @Description Check if space can use additional quota
// @Tags sys
// @Produce json
// @Param spaceId query string true "Space ID"
// @Param quota_type query string true "Quota Type"
// @Param amount query int true "Requested Amount"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/quotas/check [get]
// @Security Bearer
func (h *spaceQuotaHandler) CheckLimit(c *gin.Context) {
	spaceID := c.Query("spaceId")
	quotaType := c.Query("quota_type")
	amountStr := c.Query("amount")

	if spaceID == "" || quotaType == "" || amountStr == "" {
		resp.Fail(c.Writer, resp.BadRequest("Missing required parameters"))
		return
	}

	amount := int64(0)
	if val, err := convert.StringToInt64(amountStr); err == nil {
		amount = val
	} else {
		resp.Fail(c.Writer, resp.BadRequest("Invalid amount"))
		return
	}

	allowed, err := h.s.SpaceQuota.CheckQuotaLimit(c.Request.Context(), spaceID, structs.QuotaType(quotaType), amount)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"allowed": allowed})
}

// GetSummary handles retrieving space quota summary
//
// @Summary Get space quota summary
// @Description Retrieve all quotas for a space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {array} structs.ReadSpaceQuota "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/quotas [get]
// @Security Bearer
func (h *spaceQuotaHandler) GetSummary(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("space_id")))
		return
	}

	result, err := h.s.SpaceQuota.GetSpaceQuotaSummary(c.Request.Context(), spaceID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

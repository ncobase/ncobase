package handler

import (
	"ncobase/counter/service"
	"ncobase/counter/structs"

	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// CounterHandlerInterface represents the counter handler interface.
type CounterHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
}

// counterHandler represents the counter handler.
type counterHandler struct {
	s *service.Service
}

// NewCounterHandler creates new counter handler.
func NewCounterHandler(svc *service.Service) CounterHandlerInterface {
	return &counterHandler{
		s: svc,
	}
}

// Create handles creating a new counter.
//
// @Summary Create counter
// @Description Create a new counter.
// @Tags plug
// @Accept json
// @Produce json
// @Param body body structs.CounterBody true "CounterBody object"
// @Success 200 {object} structs.ReadCounter "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /plug/counters [post]
// @Security Bearer
func (h *counterHandler) Create(c *gin.Context) {
	body := &structs.CreateCounterBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Counter.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles updating a counter.
//
// @Summary Update counter
// @Description Update an existing counter.
// @Tags plug
// @Accept json
// @Produce json
// @Param body body structs.UpdateCounterBody true "UpdateCounterBody object"
// @Success 200 {object} structs.ReadCounter "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /plug/counters [put]
// @Security Bearer
func (h *counterHandler) Update(c *gin.Context) {
	body := types.JSON{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Counter.Update(c.Request.Context(), body["id"].(string), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles retrieving a counter by ID.
//
// @Summary Get counter
// @Description Retrieve a counter by ID.
// @Tags plug
// @Produce json
// @Param id path string true "Counter ID"
// @Success 200 {object} structs.ReadCounter "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /plug/counters/{id} [get]
// @Security Bearer
func (h *counterHandler) Get(c *gin.Context) {
	params := &structs.FindCounter{Counter: c.Param("id")}

	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Counter.Get(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles deleting a counter by ID.
//
// @Summary Delete counter
// @Description Delete a counter by ID.
// @Tags plug
// @Produce json
// @Param id path string true "Counter ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /plug/counters/{id} [delete]
// @Security Bearer
func (h *counterHandler) Delete(c *gin.Context) {
	params := &structs.FindCounter{Counter: c.Param("id")}
	err := h.s.Counter.Delete(c.Request.Context(), params.Counter)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// List handles listing all counters.
//
// @Summary List counters
// @Description Retrieve a list or tree structure of counters.
// @Tags plug
// @Produce json
// @Param params query structs.ListCounterParams true "List counter parameters"
// @Success 200 {array} structs.ReadCounter "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /plug/counters [get]
// @Security Bearer
func (h *counterHandler) List(c *gin.Context) {
	params := &structs.ListCounterParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Counter.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

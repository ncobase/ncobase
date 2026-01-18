package handler

import (
	"ncobase/plugin/payment/service"
	"ncobase/plugin/payment/structs"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"
)

// ProductHandlerInterface defines the interface for product handler operations
type ProductHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
}

// productHandler handles product-related requests
type productHandler struct {
	svc service.ProductServiceInterface
}

// NewProductHandler creates a new product handler
func NewProductHandler(svc service.ProductServiceInterface) ProductHandlerInterface {
	return &productHandler{svc: svc}
}

// Create handles the creation of a new product
//
// @Summary Create product
// @Description Create a new product
// @Tags payment,products
// @Accept json
// @Produce json
// @Param body body structs.CreateProductInput true "Product data"
// @Success 200 {object} resp.Exception{data=structs.Product} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/products [post]
// @Security Bearer
func (h *productHandler) Create(c *gin.Context) {
	var input structs.CreateProductInput
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, &input); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	// Create product
	product, err := h.svc.Create(c.Request.Context(), &input)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to create product: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to create product", err))
		return
	}

	resp.Success(c.Writer, product)
}

// Get handles getting a product by ID
//
// @Summary Get product
// @Description Get a product by ID
// @Tags payment,products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} resp.Exception{data=structs.Product} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 404 {object} resp.Exception "not found"
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/products/{id} [get]
// @Security Bearer
func (h *productHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	// Get product
	product, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to get product: %v", err)
		resp.Fail(c.Writer, resp.NotFound("Product not found"))
		return
	}

	resp.Success(c.Writer, product)
}

// Update handles updating a product
//
// @Summary Update product
// @Description Update an existing product
// @Tags payment,products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param body body structs.UpdateProductInput true "Updates"
// @Success 200 {object} resp.Exception{data=structs.Product} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/products/{id} [put]
// @Security Bearer
func (h *productHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	var input structs.UpdateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid request data", err))
		return
	}

	// Set ID from path parameter
	input.ID = id

	// Update product
	product, err := h.svc.Update(c.Request.Context(), id, &input)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to update product: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to update product", err))
		return
	}

	resp.Success(c.Writer, product)
}

// Delete handles deleting a product
//
// @Summary Delete product
// @Description Delete a product by ID
// @Tags payment,products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} resp.Exception{data=nil} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/products/{id} [delete]
// @Security Bearer
func (h *productHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	// Delete product
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		logger.Errorf(c.Request.Context(), "Failed to delete product: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to delete product", err))
		return
	}

	resp.Success(c.Writer, nil)
}

// List handles listing products
//
// @Summary List products
// @Description Get a paginated list of products
// @Tags payment,products
// @Produce json
// @Param status query string false "Filter by status"
// @Param pricing_type query string false "Filter by pricing type"
// @Param space_id query string false "Filter by space ID"
// @Param cursor query string false "Cursor for pagination"
// @Param page_size query int false "Page size" default(20)
// @Success 200 {array} structs.Product "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/products [get]
// @Security Bearer
func (h *productHandler) List(c *gin.Context) {
	var query structs.ProductQuery
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, &query); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	// Set default page size if not provided
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	// List products
	result, err := h.svc.List(c.Request.Context(), &query)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to list products: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to list products", err))
		return
	}

	resp.Success(c.Writer, result)
}

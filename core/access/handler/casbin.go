package handler

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"ncobase/core/access/service"
	"ncobase/core/access/structs"
	"net/http"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// CasbinHandlerInterface is the interface for the handler.
type CasbinHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	GetByRule(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	BulkCreate(c *gin.Context)
	Import(c *gin.Context)
	Export(c *gin.Context)
	Validate(c *gin.Context)
}

// casbinHandler represents the handler.
type casbinHandler struct {
	s *service.Service
}

// NewCasbinHandler creates a new handler.
func NewCasbinHandler(svc *service.Service) CasbinHandlerInterface {
	return &casbinHandler{
		s: svc,
	}
}

type casbinBulkCreateBody struct {
	Policies []*structs.CasbinRuleBody `json:"policies"`
}

// Create handles the creation of a Casbin rule.
//
// @Summary Create Casbin rule
// @Description Create a new Casbin rule.
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.CasbinRuleBody true "CasbinRuleBody object"
// @Success 200 {object} structs.ReadCasbinRule "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/policies [post]
// @Security Bearer
func (h *casbinHandler) Create(c *gin.Context) {
	body := &structs.CasbinRuleBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}
	if err := validateCasbinPolicy(body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.s.Casbin.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating a Casbin rule.
//
// @Summary Update Casbin rule
// @Description Update an existing Casbin rule, either fully or partially.
// @Tags sys
// @Accept json
// @Produce json
// @Param id path string true "Casbin rule ID"
// @Param body body structs.CasbinRuleBody true "CasbinRuleBody object"
// @Success 200 {object} structs.ReadCasbinRule "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/policies/{id} [put]
// @Security Bearer
func (h *casbinHandler) Update(c *gin.Context) {
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

	result, err := h.s.Casbin.Update(c.Request.Context(), id, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles getting a Casbin rule.
//
// @Summary Get Casbin rule
// @Description Retrieve details of a Casbin rule.
// @Tags sys
// @Produce json
// @Param id path string true "Casbin rule ID"
// @Success 200 {object} structs.ReadCasbinRule "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/policies/{id} [get]
// @Security Bearer
func (h *casbinHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.s.Casbin.Get(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetByRule handles finding a Casbin rule by rule components.
func (h *casbinHandler) GetByRule(c *gin.Context) {
	params := &structs.ListCasbinRuleParams{
		PType: queryPointer(c, "p_type"),
		V0:    queryPointer(c, "v0"),
		V1:    queryPointer(c, "v1"),
		V2:    queryPointer(c, "v2"),
		V3:    queryPointer(c, "v3"),
		V4:    queryPointer(c, "v4"),
		V5:    queryPointer(c, "v5"),
		Limit: 1,
	}

	result, err := h.s.Casbin.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	if len(result.Items) == 0 {
		resp.Fail(c.Writer, resp.NotFound("policy not found"))
		return
	}

	resp.Success(c.Writer, result.Items[0])
}

// Delete handles deleting a Casbin rule.
//
// @Summary Delete Casbin rule
// @Description Delete an existing Casbin rule.
// @Tags sys
// @Produce json
// @Param id path string true "Casbin rule ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/policies/{id} [delete]
// @Security Bearer
func (h *casbinHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.s.Casbin.Delete(c.Request.Context(), id); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing Casbin rules.
//
// @Summary List Casbin rules
// @Description Retrieve a list of Casbin rules.
// @Tags sys
// @Produce json
// @Param params query structs.ListCasbinRuleParams true "ListCasbinRuleParams object"
// @Success 200 {array} structs.CasbinRuleBody "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/policies [get]
// @Security Bearer
func (h *casbinHandler) List(c *gin.Context) {
	params := &structs.ListCasbinRuleParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	casbinRules, err := h.s.Casbin.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, casbinRules)
}

// BulkCreate handles creating multiple Casbin rules.
func (h *casbinHandler) BulkCreate(c *gin.Context) {
	body := &casbinBulkCreateBody{}
	if err := c.ShouldBindJSON(body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	if len(body.Policies) == 0 {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("policies")))
		return
	}

	results := make([]*structs.ReadCasbinRule, 0, len(body.Policies))
	for _, policy := range body.Policies {
		if err := validateCasbinPolicy(policy); err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		result, err := h.s.Casbin.Create(c.Request.Context(), policy)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		results = append(results, result)
	}

	resp.Success(c.Writer, results)
}

// Import handles importing Casbin rules from an array or a {policies: []} payload.
func (h *casbinHandler) Import(c *gin.Context) {
	var payload any
	if err := c.ShouldBindJSON(&payload); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	policies, err := decodeCasbinPolicies(payload)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	if len(policies) == 0 {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("policies")))
		return
	}

	results := make([]*structs.ReadCasbinRule, 0, len(policies))
	for _, policy := range policies {
		if err := validateCasbinPolicy(policy); err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		result, err := h.s.Casbin.Create(c.Request.Context(), policy)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		results = append(results, result)
	}

	resp.Success(c.Writer, gin.H{"items": results, "total": len(results)})
}

// Export handles exporting Casbin rules as JSON or CSV.
func (h *casbinHandler) Export(c *gin.Context) {
	params := &structs.ListCasbinRuleParams{Limit: 10000}
	result, err := h.s.Casbin.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	if c.DefaultQuery("format", "json") == "csv" {
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", `attachment; filename="casbin_policies.csv"`)
		c.Status(http.StatusOK)

		writer := csv.NewWriter(c.Writer)
		_ = writer.Write([]string{"id", "p_type", "v0", "v1", "v2", "v3", "v4", "v5"})
		for _, policy := range result.Items {
			_ = writer.Write([]string{
				policy.ID,
				policy.PType,
				policy.V0,
				policy.V1,
				policy.V2,
				valueOrEmpty(policy.V3),
				valueOrEmpty(policy.V4),
				valueOrEmpty(policy.V5),
			})
		}
		writer.Flush()
		return
	}

	resp.Success(c.Writer, result.Items)
}

// Validate handles validating a Casbin rule body without persisting it.
func (h *casbinHandler) Validate(c *gin.Context) {
	body := &structs.CasbinRuleBody{}
	if err := c.ShouldBindJSON(body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	if err := validateCasbinPolicy(body); err != nil {
		resp.Success(c.Writer, gin.H{"valid": false, "error": err.Error()})
		return
	}

	resp.Success(c.Writer, gin.H{"valid": true})
}

func queryPointer(c *gin.Context, key string) *string {
	value := c.Query(key)
	if value == "" {
		return nil
	}
	return &value
}

func decodeCasbinPolicies(payload any) ([]*structs.CasbinRuleBody, error) {
	if body, ok := payload.(map[string]any); ok {
		payload = body["policies"]
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var policies []*structs.CasbinRuleBody
	if err := json.Unmarshal(raw, &policies); err != nil {
		return nil, err
	}
	return policies, nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func validateCasbinPolicy(policy *structs.CasbinRuleBody) error {
	if policy == nil {
		return errors.New(ecode.FieldIsRequired("policy"))
	}
	if policy.PType == "" {
		return errors.New(ecode.FieldIsRequired("p_type"))
	}
	if policy.V0 == "" {
		return errors.New(ecode.FieldIsRequired("v0"))
	}
	if policy.V1 == "" {
		return errors.New(ecode.FieldIsRequired("v1"))
	}
	if policy.PType == "p" && policy.V2 == "" {
		return errors.New(ecode.FieldIsRequired("v2"))
	}
	return nil
}

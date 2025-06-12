package handler

import (
	"ncobase/user/service"
	"ncobase/user/structs"

	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// EmployeeHandlerInterface defines the employee handler interface
type EmployeeHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	GetByDepartment(c *gin.Context)
	GetByManager(c *gin.Context)
}

// employeeHandler implements the EmployeeHandlerInterface
type employeeHandler struct {
	s *service.Service
}

// NewEmployeeHandler creates a new employee handler
func NewEmployeeHandler(svc *service.Service) EmployeeHandlerInterface {
	return &employeeHandler{
		s: svc,
	}
}

// Create handles creating a new employee record
//
// @Summary Create employee
// @Description Create a new employee record
// @Tags sys
// @Accept json
// @Produce json
// @Param employee body structs.CreateEmployeeBody true "Employee information"
// @Success 200 {object} structs.ReadEmployee "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/employees [post]
func (h *employeeHandler) Create(c *gin.Context) {
	var body structs.CreateEmployeeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.s.Employee.Create(c.Request.Context(), &body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles updating an employee record
//
// @Summary Update employee
// @Description Update an existing employee record
// @Tags sys
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param employee body structs.UpdateEmployeeBody true "Employee information to update"
// @Success 200 {object} structs.ReadEmployee "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/employees/{user_id} [put]
func (h *employeeHandler) Update(c *gin.Context) {
	var body structs.UpdateEmployeeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	userID := c.Param("user_id")
	result, err := h.s.Employee.Update(c.Request.Context(), userID, &body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles retrieving an employee record
//
// @Summary Get employee
// @Description Retrieve an employee record by user ID
// @Tags sys
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} structs.ReadEmployee "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/employees/{user_id} [get]
func (h *employeeHandler) Get(c *gin.Context) {
	userID := c.Param("user_id")
	result, err := h.s.Employee.Get(c.Request.Context(), userID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles deleting an employee record
//
// @Summary Delete employee
// @Description Delete an employee record
// @Tags sys
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/employees/{user_id} [delete]
func (h *employeeHandler) Delete(c *gin.Context) {
	userID := c.Param("user_id")
	err := h.s.Employee.Delete(c.Request.Context(), userID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, nil)
}

// List handles listing employees
//
// @Summary List employees
// @Description List employees with filtering and pagination
// @Tags sys
// @Produce json
// @Param space_id query string false "Space ID"
// @Param department query string false "Department"
// @Param status query string false "Employee status"
// @Param employment_type query string false "Employment type"
// @Param manager_id query string false "Manager ID"
// @Param cursor query string false "Cursor for pagination"
// @Param limit query int false "Number of items to return"
// @Param direction query string false "Direction of pagination"
// @Success 200 {array} structs.ReadEmployee "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/employees [get]
func (h *employeeHandler) List(c *gin.Context) {
	params := &structs.ListEmployeeParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Employee.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetByDepartment handles retrieving employees by department
//
// @Summary Get employees by department
// @Description Retrieve all employees in a specific department
// @Tags sys
// @Produce json
// @Param department path string true "Department name"
// @Success 200 {array} structs.ReadEmployee "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/employees/department/{department} [get]
func (h *employeeHandler) GetByDepartment(c *gin.Context) {
	department := c.Param("department")
	result, err := h.s.Employee.GetByDepartment(c.Request.Context(), department)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetByManager handles retrieving employees by manager
//
// @Summary Get employees by manager
// @Description Retrieve all employees under a specific manager
// @Tags sys
// @Produce json
// @Param manager_id path string true "Manager ID"
// @Success 200 {array} structs.ReadEmployee "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/employees/manager/{manager_id} [get]
func (h *employeeHandler) GetByManager(c *gin.Context) {
	managerID := c.Param("manager_id")
	result, err := h.s.Employee.GetByManager(c.Request.Context(), managerID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

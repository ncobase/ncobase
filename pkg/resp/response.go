package resp

import (
	"net/http"
	"stocms/pkg/ecode"
	"stocms/pkg/types"

	"github.com/gin-gonic/gin"
)

// Exception represents the response structure.
type Exception struct {
	Status  int         `json:"status,omitempty"`  // HTTP status
	Code    int         `json:"code,omitempty"`    // Business code
	Message string      `json:"message,omitempty"` // Message
	Data    interface{} `json:"data,omitempty"`    // Response data
}

// response builds the response structure.
func response(code int, message string, data interface{}) *Exception {
	return &Exception{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// fail builds the failure response.
func fail(r *Exception) (int, interface{}) {
	status := http.StatusBadRequest
	code := ecode.RequestErr
	message := ecode.Text(code)

	if r.Status != 0 {
		status = r.Status
	}
	if r.Code != 0 {
		code = r.Code
	}
	if r.Message != "" {
		message = r.Message
	}

	return status, response(code, message, nil)
}

// success builds the success response.
func success(r *Exception) (int, interface{}) {
	status := http.StatusOK

	if r != nil {
		if r.Status != 0 {
			status = r.Status
		}
		if status < 200 || status >= 400 {
			return fail(r)
		}
		return status, r.Data
	}

	return status, types.JSON{"message": "ok"}
}

// write writes the response based on the specified type.
func write(context *gin.Context, contextType string, code int, res interface{}) {
	switch contextType {
	case "IndentedJSON":
		context.IndentedJSON(code, res)
	case "SecureJSON":
		context.SecureJSON(code, res)
	case "JSON":
		context.JSON(code, res)
	case "AsciiJSON":
		context.AsciiJSON(code, res)
	case "PureJSON":
		context.PureJSON(code, res)
	case "XML":
		context.XML(code, res)
	case "YAML":
		context.YAML(code, res)
	case "ProtoBuf":
		context.ProtoBuf(code, res)
	}
	context.Abort()
}

// Fail handles failure responses.
func Fail(context *gin.Context, r *Exception) {
	contextType := "JSON"
	statusCode, result := fail(r)
	write(context, contextType, statusCode, result)
}

// Success handles success responses.
func Success(context *gin.Context, r *Exception) {
	contextType := "JSON"
	statusCode, result := success(r)
	write(context, contextType, statusCode, result)
}

package resp

import (
	"encoding/json"
	"net/http"
	"stocms/pkg/ecode"
	"stocms/pkg/types"
)

// Exception represents the response structure.
type Exception struct {
	Status  int    `json:"status,omitempty"`  // HTTP status
	Code    int    `json:"code,omitempty"`    // Business code
	Message string `json:"message,omitempty"` // Message
	Data    any    `json:"data,omitempty"`    // Response data
}

// response builds the response structure.
func response(code int, message string, data any) *Exception {
	return &Exception{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// fail builds the failure response.
func fail(r *Exception) (int, any) {
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
func success(r *Exception) (int, any) {
	status := http.StatusOK

	if r != nil && r.Status != 0 {
		status = r.Status
	}

	if status < 200 || status >= 400 {
		return fail(r)
	}

	if r != nil && r.Data != nil {
		return status, r.Data
	}

	return status, types.JSON{"message": "ok"}
}

// write writes the response based on the specified type and status code.
func write(w http.ResponseWriter, contextType string, code int, res any) {
	w.WriteHeader(code)
	switch contextType {
	case "JSON":
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			return
		}
	case "XML":
		w.Header().Set("Content-Type", "application/xml")
		// Implement XML encoding here
	case "Text":
		w.Header().Set("Content-Type", "text/plain")
		// Convert res to string if needed and write it to response writer
	default:
		// Default to JSON if no contextType matches
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			return
		}
	}
}

// Fail handles failure responses.
func Fail(w http.ResponseWriter, r *Exception) {
	contextType := "JSON"
	statusCode, result := fail(r)
	write(w, contextType, statusCode, result)
}

// Success handles success responses.
func Success(w http.ResponseWriter, r *Exception) {
	contextType := "JSON"
	statusCode, result := success(r)
	write(w, contextType, statusCode, result)
}

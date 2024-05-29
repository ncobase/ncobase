package resp

import (
	"net/http"
	"stocms/pkg/ecode"
	"stocms/pkg/types"
)

// newResponse creates a new response.
func newResponse(status int, code int, message string, data ...types.JSON) *Exception {
	var responseData interface{}
	if len(data) > 0 {
		responseData = data[0]
	}
	return &Exception{
		Status:  status,
		Code:    code,
		Message: message,
		Data:    responseData,
	}
}

// AlreadyExists indicates that the resource already exists.
func AlreadyExists(message string, data ...types.JSON) *Exception {
	return newResponse(http.StatusOK, 2003, message, data...)
}

// NotExists indicates that the resource does not exist.
func NotExists(message string, data ...types.JSON) *Exception {
	return newResponse(http.StatusOK, 2002, message, data...)
}

// DBQuery indicates a database query error.
func DBQuery(message string, data ...types.JSON) *Exception {
	return newResponse(http.StatusInternalServerError, ecode.ServerErr, message, data...)
}

// Transactions indicates a transaction processing failure.
func Transactions(message string, data ...types.JSON) *Exception {
	return newResponse(http.StatusInternalServerError, ecode.ServerErr, message, data...)
}

// UnAuthorized indicates that the request is unauthorized.
func UnAuthorized(message string, data ...types.JSON) *Exception {
	return newResponse(http.StatusUnauthorized, ecode.Unauthorized, message, data...)
}

// BadRequest indicates a bad request.
func BadRequest(message string, data ...types.JSON) *Exception {
	return newResponse(http.StatusBadRequest, ecode.RequestErr, message, data...)
}

// NotFound indicates that the requested resource is not found.
func NotFound(message string, data ...types.JSON) *Exception {
	return newResponse(http.StatusNotFound, ecode.NothingFound, message, data...)
}

// Forbidden indicates access is forbidden.
func Forbidden(message string, data ...types.JSON) *Exception {
	return newResponse(http.StatusForbidden, ecode.AccessDenied, message, data...)
}

// InternalServer indicates a server error.
func InternalServer(message string, data ...types.JSON) *Exception {
	return newResponse(http.StatusInternalServerError, ecode.ServerErr, message, data...)
}

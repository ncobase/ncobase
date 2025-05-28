package types

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorCode represents a type for error codes.
// Each error code uniquely identifies a specific error condition.
type ErrorCode string

// Error returns the error code as a string.
func (e ErrorCode) Error() string {
	return string(e)
}

const (
	// General errors

	ErrInvalidParam   ErrorCode = "INVALID_PARAM"   // Invalid input parameters.
	ErrNotFound       ErrorCode = "NOT_FOUND"       // Requested resource was not found.
	ErrTimeout        ErrorCode = "TIMEOUT"         // Operation timed out.
	ErrCancelled      ErrorCode = "CANCELLED"       // Operation was cancelled.
	ErrNotImplemented ErrorCode = "NOT_IMPLEMENTED" // Functionality is not implemented.
	ErrValidation     ErrorCode = "VALIDATION"      // Validation error occurred.
	ErrPermission     ErrorCode = "PERMISSION"      // Permission denied for the operation.
	ErrConflict       ErrorCode = "CONFLICT"        // Resource conflict occurred.
	ErrUnknown        ErrorCode = "UNKNOWN"         // Unknown error occurred.
	ErrInvalidStatus  ErrorCode = "INVALID_STATUS"  // Invalid status.
	ErrNotRunning     ErrorCode = "NOT_RUNNING"     // Not running.
	ErrNotSupported   ErrorCode = "NOT_SUPPORTED"   // Not supported.

	// Execution-related errors

	ErrExecutionFailed  ErrorCode = "EXECUTION_FAILED"  // Workflow execution failed.
	ErrValidationFailed ErrorCode = "VALIDATION_FAILED" // Validation during execution failed.
	ErrRetryExceeded    ErrorCode = "RETRY_EXCEEDED"    // Maximum retry attempts exceeded.
	ErrNonRetryable     ErrorCode = "NON_RETRYABLE"     // Error is not eligible for retry.
	ErrDependencyFailed ErrorCode = "DEPENDENCY_FAILED" // Dependency service or task failed.

	// Resource errors

	ErrResourceExhausted ErrorCode = "RESOURCE_EXHAUSTED" // Resource limit exceeded.
	ErrResourceLocked    ErrorCode = "RESOURCE_LOCKED"    // Resource is locked or unavailable.
	ErrQuotaExceeded     ErrorCode = "QUOTA_EXCEEDED"     // Quota for the operation exceeded.

	// System errors

	ErrSystem             ErrorCode = "SYSTEM_ERROR"        // General system error.
	ErrInternal           ErrorCode = "INTERNAL_ERROR"      // Critical internal error.
	ErrDataCorruption     ErrorCode = "DATA_CORRUPTION"     // Data corruption detected.
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE" // Service is unavailable.
	ErrRateLimit          ErrorCode = "RATE_LIMIT"          // Rate limit exceeded.

	// Network-related errors

	ErrNetwork    ErrorCode = "NETWORK_ERROR" // Network connectivity issue.
	ErrConnection ErrorCode = "CONNECTION"    // Connection failed or closed.
	ErrProtocol   ErrorCode = "PROTOCOL"      // Protocol mismatch or violation.

	// User or authentication errors

	ErrAuthentication ErrorCode = "AUTHENTICATION"  // Authentication failed.w
	ErrUnauthorized   ErrorCode = "UNAUTHORIZED"    // User is not authorized.
	ErrSessionExpired ErrorCode = "SESSION_EXPIRED" // User session has expired.

	// Configuration errors

	ErrInvalidConfig ErrorCode = "INVALID_CONFIG" // Configuration is invalid or incomplete.
	ErrMissingConfig ErrorCode = "MISSING_CONFIG" // Required configuration is missing.
)

// WorkflowError represents a structured error in the workflow system.
// It provides detailed information about the error, including its code, message, cause, and optional metadata.
type WorkflowError struct {
	Code    ErrorCode      // A unique code representing the type of error.
	Message string         // A descriptive message about the error.
	Cause   error          // The original error that caused this error, if any.
	Details map[string]any // Additional details or metadata about the error.
	Stack   string         // A stack trace for debugging purposes.
}

// Error returns a concise string representation of the WorkflowError.
//
// Format: "[ERROR_CODE] Error message"
func (e *WorkflowError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// FullError provides a detailed string representation of the WorkflowError.
//
// Includes the error code, message, cause (if present), and any additional details.
//
// Example:
//
//	[ERROR_CODE] Error message: caused by Original cause (details: key1=value1, key2=value2)
func (e *WorkflowError) FullError() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("[%s] %s", e.Code, e.Message))
	if e.Cause != nil {
		b.WriteString(fmt.Sprintf(": caused by %s", e.Cause.Error()))
	}
	if len(e.Details) > 0 {
		b.WriteString(" (details: ")
		first := true
		for k, v := range e.Details {
			if !first {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%s=%v", k, v))
			first = false
		}
		b.WriteString(")")
	}
	return b.String()
}

// NewError creates and returns a new WorkflowError instance.
//
// Parameters:
//   - code: The error code representing the type of error.
//   - message: A descriptive message about the error.
//   - cause: (Optional) The original error that caused this error.
//
// Returns:
//
//	A pointer to a WorkflowError instance.
//
// Example usage:
//
//	err := NewError(ErrInvalidParam, "Invalid user input", nil)
func NewError(code ErrorCode, message string, cause error) *WorkflowError {
	return &WorkflowError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Details: make(map[string]any),
		Stack:   getStackTrace(),
	}
}

// AddDetail adds additional metadata to the WorkflowError.
//
// Parameters:
//   - key: The key for the detail.
//   - value: The value for the detail.
//
// Example usage:
//
//	err.AddDetail("field", "username")
func (e *WorkflowError) AddDetail(key string, value any) {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
}

// WrapError creates a WorkflowError by wrapping an existing error.
//
// Parameters:
//   - cause: The original error to wrap.
//   - code: The error code for the new WorkflowError.
//   - message: A descriptive message about the new error.
//
// Returns:
//
//	A pointer to a WorkflowError instance.
//
// Example usage:
//
//	originalErr := fmt.Errorf("database timeout")
//	err := WrapError(originalErr, ErrTimeout, "Failed to retrieve data")
func WrapError(cause error, code ErrorCode, message string) *WorkflowError {
	return &WorkflowError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Details: make(map[string]any),
		Stack:   getStackTrace(),
	}
}

// Helper function to get a mock stack trace for debugging purposes.
// Replace this with actual stack trace generation if required.
func getStackTrace() string {
	// Placeholder implementation for stack trace generation.
	return "mock-stack-trace"
}

// IsErrorCode checks if the given error matches any of the specified error codes.
//
// Parameters:
//   - err: The error to check.
//   - codes: A list of error codes to compare against.
//
// Returns:
//
//	True if the error code matches one of the provided codes, false otherwise.
//
// Example usage:
//
//	if IsErrorCode(err, ErrTimeout, ErrNetworkError) {
//	    fmt.Println("Operation failed due to timeout or network issues")
//	}
func IsErrorCode(err error, codes ...ErrorCode) bool {
	var wErr *WorkflowError
	if !errors.As(err, &wErr) {
		return false
	}
	for _, code := range codes {
		if errors.Is(code, wErr.Code) {
			return true
		}
	}
	return false
}

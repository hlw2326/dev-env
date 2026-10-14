package errors

import (
	"fmt"
)

// Standard error codes matching API contract
const (
	CodeInvalidArgument     = "INVALID_ARGUMENT"
	CodeValidationFailed    = "VALIDATION_FAILED"
	CodeNotFound            = "NOT_FOUND"
	CodeAlreadyExists       = "ALREADY_EXISTS"
	CodePathInvalid         = "PATH_INVALID"
	CodePathNotAccessible   = "PATH_NOT_ACCESSIBLE"
	CodePortAlreadyInUse    = "PORT_ALREADY_IN_USE"
	CodePHPVersionInUse     = "PHP_VERSION_IN_USE"
	CodeRuntimeNotInstalled = "RUNTIME_NOT_INSTALLED"
	CodeProcessStartFailed  = "PROCESS_START_FAILED"
	CodeProcessStopFailed   = "PROCESS_STOP_FAILED"
	CodeConfigInvalid       = "CONFIG_INVALID"
	CodePermissionDenied    = "PERMISSION_DENIED"
	CodeDownloadFailed      = "DOWNLOAD_FAILED"
	CodeInstallFailed       = "INSTALL_FAILED"
	CodeTaskNotCancellable  = "TASK_NOT_CANCELLABLE"
	CodeServiceRunning      = "SERVICE_RUNNING"
	CodeDatabaseError       = "DATABASE_ERROR"
	CodeInternalError       = "INTERNAL_ERROR"
)

// AppError is the standard structured error for DevEnv
type AppError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Retryable bool                   `json:"retryable,omitempty"`
	Err       error                  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with a code and message
func Wrap(err error, code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// WithDetails adds details to an AppError
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// WithRetryable sets retryable flag
func (e *AppError) WithRetryable(retryable bool) *AppError {
	e.Retryable = retryable
	return e
}

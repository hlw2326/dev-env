package errors

import "fmt"

const (
	INVALID_ARGUMENT         = "INVALID_ARGUMENT"
	NOT_FOUND                = "NOT_FOUND"
	ALREADY_EXISTS           = "ALREADY_EXISTS"
	PATH_INVALID             = "PATH_INVALID"
	PATH_NOT_ACCESSIBLE      = "PATH_NOT_ACCESSIBLE"
	PORT_ALREADY_IN_USE      = "PORT_ALREADY_IN_USE"
	RUNTIME_IN_USE           = "RUNTIME_IN_USE"
	RUNTIME_NOT_INSTALLED    = "RUNTIME_NOT_INSTALLED"
	PROCESS_START_FAILED     = "PROCESS_START_FAILED"
	PERMISSION_DENIED        = "PERMISSION_DENIED"
	DOWNLOAD_FAILED          = "DOWNLOAD_FAILED"
	INSTALL_FAILED           = "INSTALL_FAILED"
	TASK_NOT_CANCELLABLE     = "TASK_NOT_CANCELLABLE"
	SERVICE_RUNNING          = "SERVICE_RUNNING"
	DATABASE_ERROR           = "DATABASE_ERROR"
	DATABASE_SCHEMA_MISMATCH = "DATABASE_SCHEMA_MISMATCH"
	MIGRATION_FAILED         = "MIGRATION_FAILED"
	GIT_NOT_INSTALLED        = "GIT_NOT_INSTALLED"
	GIT_NOT_INITIALIZED      = "GIT_NOT_INITIALIZED"
	GIT_OPERATION_FAILED     = "GIT_OPERATION_FAILED"
	GIT_NO_CHANGES           = "GIT_NO_CHANGES"
	DEPLOYMENT_FAILED        = "DEPLOYMENT_FAILED"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func Wrap(err error, code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

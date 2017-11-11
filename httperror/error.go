package httperror

import (
	"fmt"
)

var (
	INVALID_DATE_FORMAT  = ErrorCode{"InvalidDateFormat", 422}
	INVALID_FORMAT       = ErrorCode{"InvalidFormat", 422}
	INVALID_REFERENCE    = ErrorCode{"InvalidReference", 422}
	NOT_NULLABLE         = ErrorCode{"NotNullable", 422}
	NOT_UNIQUE           = ErrorCode{"NotUnique", 422}
	MIN_LIMIT_EXCEEDED   = ErrorCode{"MinLimitExceeded", 422}
	MAX_LIMIT_EXCEEDED   = ErrorCode{"MaxLimitExceeded", 422}
	MIN_LENGTH_EXCEEDED  = ErrorCode{"MinLengthExceeded", 422}
	MAX_LENGTH_EXCEEDED  = ErrorCode{"MaxLengthExceeded", 422}
	INVALID_OPTION       = ErrorCode{"InvalidOption", 422}
	INVALID_CHARACTERS   = ErrorCode{"InvalidCharacters", 422}
	MISSING_REQUIRED     = ErrorCode{"MissingRequired", 422}
	INVALID_CSRF_TOKEN   = ErrorCode{"InvalidCSRFToken", 422}
	INVALID_ACTION       = ErrorCode{"InvalidAction", 422}
	INVALID_BODY_CONTENT = ErrorCode{"InvalidBodyContent", 422}
	INVALID_TYPE         = ErrorCode{"InvalidType", 422}
	ACTION_NOT_AVAILABLE = ErrorCode{"ActionNotAvailable", 404}
	INVALID_STATE        = ErrorCode{"InvalidState", 422}
	SERVER_ERROR         = ErrorCode{"ServerError", 500}

	METHOD_NOT_ALLOWED = ErrorCode{"MethodNotAllow", 405}
	NOT_FOUND          = ErrorCode{"NotFound", 404}
)

type ErrorCode struct {
	code   string
	status int
}

func (e ErrorCode) String() string {
	return fmt.Sprintf("%s %d", e.code, e.status)
}

type APIError struct {
	code      ErrorCode
	message   string
	Cause     error
	fieldName string
}

func NewAPIError(code ErrorCode, message string) error {
	return &APIError{
		code:    code,
		message: message,
	}
}

func NewFieldAPIError(code ErrorCode, fieldName, message string) error {
	return &APIError{
		code:      code,
		message:   message,
		fieldName: fieldName,
	}
}

func WrapFieldAPIError(err error, code ErrorCode, fieldName, message string) error {
	return &APIError{
		Cause:     err,
		code:      code,
		message:   message,
		fieldName: fieldName,
	}
}

func WrapAPIError(err error, code ErrorCode, message string) error {
	return &APIError{
		code:    code,
		message: message,
		Cause:   err,
	}
}

func (a *APIError) Error() string {
	if a.fieldName != "" {
		return fmt.Sprintf("%s=%s: %s", a.fieldName, a.code, a.message)
	}
	return fmt.Sprintf("%s: %s", a.code, a.message)
}

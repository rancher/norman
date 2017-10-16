package httperror

import (
	"fmt"
)

var (
	INVALID_DATE_FORMAT  = ErrorCode("InvalidDateFormat")
	INVALID_FORMAT       = ErrorCode("InvalidFormat")
	INVALID_REFERENCE    = ErrorCode("InvalidReference")
	NOT_NULLABLE         = ErrorCode("NotNullable")
	NOT_UNIQUE           = ErrorCode("NotUnique")
	MIN_LIMIT_EXCEEDED   = ErrorCode("MinLimitExceeded")
	MAX_LIMIT_EXCEEDED   = ErrorCode("MaxLimitExceeded")
	MIN_LENGTH_EXCEEDED  = ErrorCode("MinLengthExceeded")
	MAX_LENGTH_EXCEEDED  = ErrorCode("MaxLengthExceeded")
	INVALID_OPTION       = ErrorCode("InvalidOption")
	INVALID_CHARACTERS   = ErrorCode("InvalidCharacters")
	MISSING_REQUIRED     = ErrorCode("MissingRequired")
	INVALID_CSRF_TOKEN   = ErrorCode("InvalidCSRFToken")
	INVALID_ACTION       = ErrorCode("InvalidAction")
	INVALID_BODY_CONTENT = ErrorCode("InvalidBodyContent")
	INVALID_TYPE         = ErrorCode("InvalidType")
	ACTION_NOT_AVAILABLE = ErrorCode("ActionNotAvailable")
	INVALID_STATE        = ErrorCode("InvalidState")
	SERVER_ERROR         = ErrorCode("ServerError")
)

type ErrorCode string

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
		Cause: err,
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
	return fmt.Sprintf("%s: %s", a.code, a.message)
}

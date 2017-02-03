package util

import "fmt"

const (
	// OtherError -- Other (maybe non HTTP-related) error
	OtherError = 0
	// BadRequestError -- Indicates an invalid request
	BadRequestError = 400
	// AuthenticationError -- Indicates invalid authentication credentials
	AuthenticationError = 401
	// ForbiddenError -- You're not allowed to access that resource
	ForbiddenError = 403
	// NotFoundError -- The resource that's been requested doesn't exist
	NotFoundError = 404
)

// APIError -- Custom error interface
type APIError interface {
	error
	Code() int
}

type apiErrorImpl struct {
	msg  string
	code int
}

func (e *apiErrorImpl) Error() string {
	return e.msg
}

func (e *apiErrorImpl) Code() int {
	return e.code
}

// NewAPIError -- Create an APIError instance
func NewAPIError(code int, msg ...interface{}) APIError {
	return &apiErrorImpl{code: code, msg: fmt.Sprint(msg...)}
}

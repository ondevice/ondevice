package util

import (
	"fmt"
	"net/http"
)

const (
	// OtherError -- Other (maybe non HTTP-related) error
	OtherError = 0
	// BadRequestError -- Indicates an invalid request (400)
	BadRequestError = http.StatusBadRequest
	// AuthenticationError -- Indicates invalid authentication credentials (401)
	AuthenticationError = http.StatusUnauthorized
	// ForbiddenError -- You're not allowed to access that resource (403)
	ForbiddenError = http.StatusForbidden
	// NotFoundError -- The resource that's been requested doesn't exist (404)
	NotFoundError = http.StatusNotFound
	// TooManyRequestsError -- You have been rate limited (429)
	TooManyRequestsError = http.StatusTooManyRequests
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

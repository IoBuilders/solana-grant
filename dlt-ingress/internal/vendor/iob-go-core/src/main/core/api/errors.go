package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

const (
	ErrorCodeRequestTimeout   coreerror.ErrorCode = "REQUEST_TIMEOUT"
	ErrorCodeInternal         coreerror.ErrorCode = "INTERNAL_ERROR"
	ErrorCodeResourceNotFound coreerror.ErrorCode = "RESOURCE_NOT_FOUND"
	ErrorCodeInvalidToken     coreerror.ErrorCode = "INVALID_TOKEN"
	ErrorCodeTokenNotReceived coreerror.ErrorCode = "TOKEN_NOT_RECEIVED"
)

var (
	ApiErrorResponseInternal *ErrorResponse = NewErrorResponse(
		"An unexpected error occurred. Please retry or contact support if the problem persists",
		ErrorCodeInternal,
		http.StatusInternalServerError,
	)

	ApiResponseErrorTimeout *ErrorResponse = NewErrorResponse(
		"Request timeout",
		ErrorCodeRequestTimeout,
		http.StatusGatewayTimeout,
	)
)

func handleError(err error) *ErrorResponse {
	if errors.Is(err, context.DeadlineExceeded) {
		return ApiResponseErrorTimeout
	}
	domErr, ok := errors.AsType[coreerror.DomainError](err)
	if !ok {
		return ApiErrorResponseInternal
	}
	switch {
	case errors.Is(err, coreerror.ErrNotFound):
		return NewErrorResponse(domErr.Error(), domErr.ErrorCode(), http.StatusNotFound)
	case errors.Is(err, coreerror.ErrValidation):
		return NewErrorResponse(domErr.Error(), domErr.ErrorCode(), http.StatusBadRequest)
	case errors.Is(err, coreerror.ErrConflict):
		return NewErrorResponse(domErr.Error(), domErr.ErrorCode(), http.StatusConflict)
	case errors.Is(err, coreerror.ErrUnauthorized):
		return NewErrorResponse(domErr.Error(), domErr.ErrorCode(), http.StatusUnauthorized)
	case errors.Is(err, coreerror.ErrForbidden):
		return NewErrorResponse(domErr.Error(), domErr.ErrorCode(), http.StatusForbidden)
	default:
		return ApiErrorResponseInternal
	}
}

func RespondError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	errResp := handleError(err)
	_ = c.Error(err)
	c.JSON(errResp.Status, errResp)
}

func RespondBindingError(c *gin.Context, err error) {
	RespondError(c, bindingError(err))
}

func bindingError(err error) error {
	if ves, ok := errors.AsType[validator.ValidationErrors](err); ok && len(ves) > 0 {
		parts := make([]string, 0, len(ves))
		for _, fe := range ves {
			parts = append(parts, fmt.Sprintf("%s: %s", fe.Field(), fe.Tag()))
		}
		return coreerror.NewValidationDomainError(domainerrors.ErrorCodeValidation, strings.Join(parts, "; "))
	}
	return coreerror.NewValidationDomainError(domainerrors.ErrorCodeValidation, "request: invalid body")
}

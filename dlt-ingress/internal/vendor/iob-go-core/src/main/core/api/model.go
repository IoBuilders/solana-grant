package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

type HttpAdapterRequest[T any] struct {
	c    *gin.Context
	Body *T
}

// Query
func (a *HttpAdapterRequest[T]) GetQueryParam(paramName string) string {
	return a.c.Query(paramName)
}

func (a *HttpAdapterRequest[T]) GetUUIDQueryParam(paramName string) (uuid.UUID, error) {
	rawUUID := a.GetQueryParam(paramName)
	if rawUUID == "" {
		return uuid.Nil, nil
	}
	val, err := uuid.Parse(rawUUID)
	if err != nil {
		return uuid.Nil, coreerror.NewValidationDomainError(domainerrors.ErrorCodeValidation, fmt.Sprintf("Invalid %s", paramName))
	}
	return val, nil
}

func (a *HttpAdapterRequest[T]) GetDateQueryParam(paramName string) (time.Time, error) {
	return a.GetDateWithLayoutQueryParam(paramName, time.RFC3339)
}

func (a *HttpAdapterRequest[T]) GetSimpleDateQueryParam(paramName string) (time.Time, error) {
	return a.GetDateWithLayoutQueryParam(paramName, time.DateOnly)
}

func (a *HttpAdapterRequest[T]) GetDateWithLayoutQueryParam(paramName string, layout string) (time.Time, error) {
	dateParam := a.GetQueryParam(paramName)
	var val time.Time
	var err error
	if dateParam != "" {
		val, err = parseDate(dateParam, layout)
		if err != nil {
			return val, coreerror.NewValidationDomainError(domainerrors.ErrorCodeValidation, fmt.Sprintf("Invalid %s", paramName))
		}
		return val, nil
	}
	return val, nil
}

func (a *HttpAdapterRequest[T]) GetQueryArrayParam(paramName string) []string {
	params := a.c.QueryArray(paramName)
	uniqueParams := make(map[string]struct{})

	for _, value := range params {
		for _, part := range strings.Split(value, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				uniqueParams[trimmed] = struct{}{}
			}
		}
	}

	result := make([]string, 0, len(uniqueParams))
	for param := range uniqueParams {
		result = append(result, param)
	}

	return result
}

// Params
func (a *HttpAdapterRequest[T]) GetParam(paramName string) string {
	return strings.TrimSpace(a.c.Param(paramName))
}

func (a *HttpAdapterRequest[T]) GetUUIDParam(paramName string) (uuid.UUID, error) {
	val, err := uuid.Parse(a.GetParam(paramName))
	if err != nil {
		return uuid.Nil, coreerror.NewValidationDomainError(domainerrors.ErrorCodeValidation, fmt.Sprintf("Invalid %s", paramName))
	}
	return val, nil
}

func (a *HttpAdapterRequest[T]) GetDateParam(paramName string) (time.Time, error) {
	return a.GetDateWithLayoutParam(paramName, time.RFC3339)
}

func (a *HttpAdapterRequest[T]) GetSimpleDateParam(paramName string) (time.Time, error) {
	return a.GetDateWithLayoutParam(paramName, time.DateOnly)
}

func (a *HttpAdapterRequest[T]) GetDateWithLayoutParam(paramName string, layout string) (time.Time, error) {
	dateParam := a.GetParam(paramName)
	var val time.Time
	if dateParam != "" {
		val, err := parseDate(dateParam, layout)
		if err != nil {
			return val, coreerror.NewValidationDomainError(domainerrors.ErrorCodeValidation, fmt.Sprintf("Invalid %s", paramName))
		}
		return val, nil
	}
	return val, nil
}

// Context
func (a *HttpAdapterRequest[T]) GetContextValue(key string) (string, bool) {
	raw, exists := a.c.Get(key)
	if !exists {
		return "", false
	}
	val, ok := raw.(string)
	return val, ok
}

// GetUUIDContextValue reads a uuid.UUID value set on the request context by a
// prior middleware (e.g. auth), whether stored as uuid.UUID directly or as a
// string that needs parsing (e.g. a JWT "sub" claim). Returns false if the
// key is missing or malformed — this signals a middleware wiring issue, not
// invalid client input, so callers should surface it as an internal error,
// not a validation error.
func (a *HttpAdapterRequest[T]) GetUUIDContextValue(key string) (uuid.UUID, bool) {
	raw, exists := a.c.Get(key)
	if !exists {
		return uuid.Nil, false
	}
	if val, ok := raw.(uuid.UUID); ok {
		return val, true
	}
	val, ok := raw.(string)
	if !ok {
		return uuid.Nil, false
	}
	parsed, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, false
	}
	return parsed, true
}

func parseDate(dateParam string, layout string) (time.Time, error) {
	val, err := time.Parse(layout, dateParam)
	if err != nil {
		return val, err
	}
	return val, nil
}

func (a *HttpAdapterRequest[T]) GetHeaderParam(paramName string) string {
	return strings.TrimSpace(a.c.GetHeader(paramName))
}

func (a *HttpAdapterRequest[T]) GetPaginationParams() (pagination.PaginationParams, error) {
	paginationParams, err := pagination.GetPaginationParams(a.c)
	if err != nil {
		return paginationParams, coreerror.NewValidationDomainError(domainerrors.ErrorCodeValidation, "Invalid pagination parameters")
	}
	return paginationParams, nil
}

type HttpAdapterResponse[T any] struct {
	StatusCode int
	Response   *T
}

type EmptyResponse = map[string]any

func NewHttpAdapterResponse[T any](statusCode int, response *T) *HttpAdapterResponse[T] {
	return &HttpAdapterResponse[T]{
		StatusCode: statusCode,
		Response:   response,
	}
}

func NewEmptyHttpAdapterResponse(statusCode int) *HttpAdapterResponse[EmptyResponse] {
	return &HttpAdapterResponse[EmptyResponse]{
		StatusCode: statusCode,
		Response:   &EmptyResponse{},
	}
}

type ErrorResponse struct {
	Error  string `json:"error"`
	Code   string `json:"code"`
	Status int    `json:"status"`
}

func NewErrorResponse(err string, errorCode coreerror.ErrorCode, status int) *ErrorResponse {
	return &ErrorResponse{
		Error:  err,
		Code:   string(errorCode),
		Status: status,
	}
}

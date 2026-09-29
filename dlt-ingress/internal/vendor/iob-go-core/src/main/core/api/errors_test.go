package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c, w
}

func decode(t *testing.T, body []byte) ErrorResponse {
	t.Helper()
	var er ErrorResponse
	assert.NoError(t, json.Unmarshal(body, &er))
	return er
}

func TestHandleError_StatusCodes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"not found domain error", coreerror.NewNotFoundDomainError("X_NOT_FOUND", "x not found"), http.StatusNotFound},
		{"validation domain error", coreerror.NewValidationDomainError(domainerrors.ErrorCodeValidation, "bad x"), http.StatusBadRequest},
		{"conflict domain error", coreerror.NewConflictDomainError("X_CONFLICT", "x exists"), http.StatusConflict},
		{"unauthorized domain error", coreerror.NewUnauthorizedDomainError("X_UNAUTHORIZED", "no auth"), http.StatusUnauthorized},
		{"forbidden domain error", coreerror.NewForbiddenDomainError("X_FORBIDDEN", "no perm"), http.StatusForbidden},
		{"unknown error", errors.New("boom"), http.StatusInternalServerError},
		{"nil", nil, http.StatusInternalServerError},
		{"sentinel without DomainError", coreerror.ErrNotFound, http.StatusInternalServerError},
		{"context deadline exceeded", context.DeadlineExceeded, http.StatusGatewayTimeout},
		{"wrapped deadline exceeded", fmt.Errorf("repo: %w", context.DeadlineExceeded), http.StatusGatewayTimeout},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := handleError(tt.err)
			assert.NotNil(t, resp.Status)
			assert.Equal(t, tt.want, resp.Status)
		})
	}
}

// stubDomainError is a test-only DomainError to exercise RespondError without
// depending on production typed errors. Each bounded context owns its own
// typed errors; this stub stands in for any of them.
type stubDomainError struct {
	sentinel error
	code     coreerror.ErrorCode
	msg      string
}

func (e *stubDomainError) Error() string                  { return e.msg }
func (e *stubDomainError) Unwrap() error                  { return e.sentinel }
func (e *stubDomainError) ErrorCode() coreerror.ErrorCode { return e.code }

func TestRespondError_DomainError(t *testing.T) {
	c, w := newTestContext()
	RespondError(c, &stubDomainError{
		sentinel: coreerror.ErrNotFound,
		code:     "ACTOR_NOT_FOUND",
		msg:      "actor not found",
	})

	assert.Equal(t, http.StatusNotFound, w.Code)
	body := decode(t, w.Body.Bytes())
	assert.Equal(t, "actor not found", body.Error)
	assert.Equal(t, "ACTOR_NOT_FOUND", body.Code)
}

func TestRespondError_ValidationDomainError(t *testing.T) {
	c, w := newTestContext()
	RespondError(c, &stubDomainError{
		sentinel: coreerror.ErrValidation,
		code:     domainerrors.ErrorCodeValidation,
		msg:      "field 'x' is required",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
	body := decode(t, w.Body.Bytes())
	assert.Equal(t, "field 'x' is required", body.Error)
	assert.Equal(t, "VALIDATION_ERROR", body.Code)
}

func TestRespondError_OpaqueInternalError(t *testing.T) {
	c, w := newTestContext()
	RespondError(c, errors.New("kafka unreachable"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	body := decode(t, w.Body.Bytes())
	assert.Equal(t, "INTERNAL_ERROR", body.Code)
	assert.Equal(t, "An unexpected error occurred. Please retry or contact support if the problem persists", body.Error, "internal cause must not leak to client")
}

func TestRespondError_NilNoOp(t *testing.T) {
	c, w := newTestContext()
	RespondError(c, nil)
	assert.Equal(t, http.StatusOK, w.Code) // gin default
	assert.Equal(t, "", w.Body.String())
}

func TestRespondBindingError_PlainError(t *testing.T) {
	c, w := newTestContext()
	RespondBindingError(c, errors.New("malformed json"))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	body := decode(t, w.Body.Bytes())
	assert.Equal(t, "VALIDATION_ERROR", body.Code)
	assert.Equal(t, "request: invalid body", body.Error, "framework-level errors must not leak to client")
}

func TestRespondBindingError_ValidatorErrors(t *testing.T) {
	type Body struct {
		Email string `json:"email" binding:"required"`
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	var body Body
	err := c.ShouldBindJSON(&body)
	if err == nil {
		t.Fatal("expected gin binding to fail on missing required field")
	}

	RespondBindingError(c, err)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	got := decode(t, w.Body.Bytes())
	assert.Equal(t, "VALIDATION_ERROR", got.Code)
	assert.Equal(t, "Email: required", got.Error)
}

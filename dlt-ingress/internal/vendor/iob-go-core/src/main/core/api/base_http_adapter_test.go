package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	//nolint:staticcheck // legacy bridge tests
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

func runAdapter[Rq any, Rs any](t *testing.T, a *BaseHttpAdapter[Rq, Rs]) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("{}"))
	c.Request.Header.Set("Content-Type", "application/json")
	a.Handler(c)
	return w
}

func failing[Rq any, Rs any](err error) func(context.Context, *HttpAdapterRequest[Rq]) (*HttpAdapterResponse[Rs], error) {
	return func(context.Context, *HttpAdapterRequest[Rq]) (*HttpAdapterResponse[Rs], error) {
		return nil, err
	}
}

func TestBaseHttpAdapter_DomainErrorUsesSemanticPath(t *testing.T) {
	a := NewHttpAdapter(
		failing[struct{}, struct{}](&stubDomainError{
			sentinel: coreerror.ErrNotFound,
			code:     "ACTOR_NOT_FOUND",
			msg:      "actor missing",
		}),
	)

	w := runAdapter(t, a)

	assert.Equal(t, http.StatusNotFound, w.Code)
	body := decode(t, w.Body.Bytes())
	assert.Equal(t, "ACTOR_NOT_FOUND", body.Code)
	assert.Equal(t, "actor missing", body.Error)
}

func TestBaseHttpAdapter_OpaqueErrorWithoutMapperIsSanitised(t *testing.T) {
	a := NewHttpAdapter(failing[struct{}, struct{}](errors.New("kafka unreachable")))

	w := runAdapter(t, a)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	body := decode(t, w.Body.Bytes())
	assert.Equal(t, "INTERNAL_ERROR", body.Code)
	assert.Equal(t, "An unexpected error occurred. Please retry or contact support if the problem persists", body.Error)
}

type bindingBody struct {
	Email string `json:"email" binding:"required"`
}

func runBindingAdapter[Rs any](t *testing.T, a *BaseHttpAdapter[bindingBody, Rs]) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("{}"))
	c.Request.Header.Set("Content-Type", "application/json")
	a.Handler(c)
	return w
}

func TestBaseHttpAdapter_BindingError_NewConstructor_UsesSemanticFormat(t *testing.T) {
	a := NewHttpAdapter(
		func(context.Context, *HttpAdapterRequest[bindingBody]) (*HttpAdapterResponse[struct{}], error) {
			t.Fatal("handle must not run when binding fails")
			return nil, nil
		},
	)

	w := runBindingAdapter(t, a)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	body := decode(t, w.Body.Bytes())
	assert.Equal(t, "VALIDATION_ERROR", body.Code, "new constructor opts into semantic binding errors")
	assert.Contains(t, body.Error, "Email", "validator field message must reach the client")
	assert.Contains(t, body.Error, "required")
}

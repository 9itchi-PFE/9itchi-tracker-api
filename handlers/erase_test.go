package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"9itchi-tracker-api/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type stubErase struct {
	id  string
	err *models.ApiError
}

func (s *stubErase) EraseUser(_ context.Context, id string) *models.ApiError {
	s.id = id
	return s.err
}

func TestEraseUserHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubErase{}
	r := gin.New()
	r.DELETE("/internal/users/:id", EraseUser(svc))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/internal/users/u-1", nil))
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "u-1", svc.id)
	assert.Empty(t, w.Body.String())

	svc.err = &models.ApiError{Code: http.StatusBadRequest, Error: assert.AnError}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/internal/users/nope", nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

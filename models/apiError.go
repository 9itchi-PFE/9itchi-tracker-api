package models

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ApiError struct {
	Code  int
	Error error
}

// NotImplemented is returned by every service stub until the feature is built.
func NotImplemented() *ApiError {
	return &ApiError{Code: http.StatusNotImplemented, Error: errors.New("not implemented")}
}

func FormatError(c *gin.Context, apiError *ApiError) {
	c.JSON(apiError.Code, gin.H{
		"error":  apiError.Error.Error(),
		"status": apiError.Code,
	})
}

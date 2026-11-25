package handlers

import (
	"net/http"

	"9itchi-tracker-api/models"
	"9itchi-tracker-api/services"

	"github.com/gin-gonic/gin"
)

// EraseUser: DELETE /internal/users/:id (internal) removes what this service keeps about a deleted account.
func EraseUser(svc services.EraseService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiErr := svc.EraseUser(c.Request.Context(), c.Param("id")); apiErr != nil {
			models.FormatError(c, apiErr)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

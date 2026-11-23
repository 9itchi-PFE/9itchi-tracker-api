package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"9itchi-tracker-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func authEngine(t *testing.T) *gin.Engine {
	t.Setenv("ITCHI_SECRET_KEY", strings.Repeat("s", 32))
	t.Setenv("ITCHI_ENCRYPTION_KEY", strings.Repeat("e", 32))
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/me", AuthMiddleware(), func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"id": c.GetString("user_id")}) })
	r.GET("/admin", AuthMiddleware(), RequireRole("admin"), func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func tokenFor(t *testing.T, role, typ string) string {
	tok, err := utils.IssueToken(utils.LoadConfig(), utils.Claims{ID: "u1", Role: role, Type: typ}, time.Minute)
	require.NoError(t, err)
	return tok
}

func do(r *gin.Engine, path, auth string) int {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestAuthMiddleware(t *testing.T) {
	r := authEngine(t)
	assert.Equal(t, http.StatusUnauthorized, do(r, "/me", ""))
	assert.Equal(t, http.StatusUnauthorized, do(r, "/me", "Bearer nope"))
	assert.Equal(t, http.StatusUnauthorized, do(r, "/me", "Bearer "+tokenFor(t, "candidate", utils.TokenRefresh)))
	assert.Equal(t, http.StatusOK, do(r, "/me", "Bearer "+tokenFor(t, "candidate", utils.TokenAccess)))
}

func TestRequireRole(t *testing.T) {
	r := authEngine(t)
	assert.Equal(t, http.StatusForbidden, do(r, "/admin", "Bearer "+tokenFor(t, "candidate", utils.TokenAccess)))
	assert.Equal(t, http.StatusOK, do(r, "/admin", "Bearer "+tokenFor(t, "admin", utils.TokenAccess)))
}

package auth_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/thornhall/simple-go-service/internal/middleware/auth"
)

func fakeProtectedHandler(c *gin.Context) {
	if userId, exists := c.Get("userId"); exists {
		c.JSON(http.StatusOK, gin.H{"got": userId})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no userObjectId found"})
	}
}

func TestJWTAuth_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	jwtSecret := os.Getenv("JWT_SECRET")
	assert.NotZero(t, jwtSecret)
	if jwtSecret == "" {
		return
	}
	jwtSecretBytes := []byte(jwtSecret)
	r.GET("/protected",
		auth.JWTAuth(jwtSecretBytes),
		fakeProtectedHandler,
	)

	// Case A: Missing or malformed Authorization header → expect 401
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"missing or invalid Authorization header"`)

	// Case B: Badly formatted “Bearer” header → still 401
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearertoken-without-space")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	//Case C: Valid “Bearer <token>,” but token signed with wrong secret → 401
	os.Setenv("JWT_SECRET", "badSecret")
	wrongToken, err := auth.IssueJWT(1, "thornhall@gmail.com")
	os.Setenv("JWT_SECRET", jwtSecret)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+wrongToken)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Case E: Valid “Bearer <token>,” correct secret, with “sub” claim → 200 + context set
	validToken, err := auth.IssueJWT(1, "thornhall@gmail.com")
	assert.NoError(t, err)
	if err != nil {
		return
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"got":1`)
}

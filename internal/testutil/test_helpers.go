// internal/testutil/test_helpers.go
package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// SetUpRouter returns a configured Gin router for testing
func SetUpRouter() *gin.Engine {
	router := gin.Default()
	gin.SetMode(gin.TestMode)
	return router
}

// MakeRequest performs an HTTP request and returns the response
func MakeRequest(method, url string, body interface{}, router *gin.Engine) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBytes)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

// AssertJSONResponse checks that the HTTP response matches expected JSON
func AssertJSONResponse(t *testing.T, resp *httptest.ResponseRecorder, expectedCode int, expectedBody interface{}) {
	assert.Equal(t, expectedCode, resp.Code)

	if expectedBody != nil {
		var actualBody interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &actualBody)
		assert.NoError(t, err)
		assert.Equal(t, expectedBody, actualBody)
	}
}

// CreateAuthToken creates a test JWT token for authentication tests
func CreateAuthToken(userID uint) string {
	// Implement token creation logic for testing
	// This depends on your JWT implementation
	return "test-jwt-token"
}

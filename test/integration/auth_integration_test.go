// test/integration/auth_integration_test.go
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aquibsayyed9/sentinel/internal/auth"
	"github.com/aquibsayyed9/sentinel/internal/config"
	"github.com/aquibsayyed9/sentinel/internal/handlers"
	"github.com/aquibsayyed9/sentinel/internal/repository"
	"github.com/aquibsayyed9/sentinel/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuthIntegrationTestSuite struct {
	suite.Suite
	router       *gin.Engine
	userRepo     repository.UserRepository
	userService  services.UserService
	tokenService auth.TokenService
	cfg          *config.Config
}

func (s *AuthIntegrationTestSuite) SetupSuite() {
	// Use test database
	db := GetTestDB()
	if db == nil {
		s.T().Fatal("Database connection is nil in SetupSuite")
	}
	s.userRepo = repository.NewUserRepository(db)
	s.userService = services.NewUserService(s.userRepo)

	// Create test config
	s.cfg = &config.Config{
		JWT: struct {
			Secret     string `mapstructure:"secret"`
			ExpireHour int    `mapstructure:"expire_hour"`
		}{
			Secret:     "test-secret-key",
			ExpireHour: 24,
		},
	}
	s.tokenService = auth.NewTokenService(s.cfg)

	// Setup router
	gin.SetMode(gin.TestMode)
	s.router = gin.Default()

	// Create auth handler
	authHandler := handlers.NewAuthHandler(s.userService, s.tokenService)

	// Set up routes
	authGroup := s.router.Group("/api/v1/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
	}

	// Protected routes (with auth middleware)
	protected := s.router.Group("/api/v1")
	protected.Use(auth.AuthMiddleware(s.tokenService))
	{
		protected.GET("/profile", authHandler.GetProfile)
	}
}

// test/integration/auth_integration_test.go (continued)
func TestAuthIntegrationSuite(t *testing.T) {
	suite.Run(t, new(AuthIntegrationTestSuite))
}

func (s *AuthIntegrationTestSuite) TestRegisterLoginFlow() {
	// Test registration
	registerBody := map[string]interface{}{
		"email":      "integration@example.com",
		"password":   "password123",
		"first_name": "Integration",
		"last_name":  "Test",
	}
	registerJSON, _ := json.Marshal(registerBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(registerJSON))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	var registerResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &registerResponse)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, w.Code)
	assert.NotNil(s.T(), registerResponse["token"])
	assert.NotNil(s.T(), registerResponse["user"])

	// Test login with the same credentials
	loginBody := map[string]interface{}{
		"email":    "integration@example.com",
		"password": "password123",
	}
	loginJSON, _ := json.Marshal(loginBody)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	var loginResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &loginResponse)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, w.Code)
	assert.NotNil(s.T(), loginResponse["token"])
	assert.NotNil(s.T(), loginResponse["user"])

	// Extract token for profile access
	token := loginResponse["token"].(string)

	// Test accessing profile with the token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	s.router.ServeHTTP(w, req)

	var profileResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &profileResponse)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, w.Code)
	assert.NotNil(s.T(), profileResponse["user"])

	user := profileResponse["user"].(map[string]interface{})
	assert.Equal(s.T(), "integration@example.com", user["email"])
	assert.Equal(s.T(), "Integration", user["first_name"])
	assert.Equal(s.T(), "Test", user["last_name"])
}

func (s *AuthIntegrationTestSuite) TestRegister_DuplicateEmail() {
	// First registration
	registerBody := map[string]interface{}{
		"email":      "duplicate@example.com",
		"password":   "password123",
		"first_name": "Duplicate",
		"last_name":  "Test",
	}
	registerJSON, _ := json.Marshal(registerBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(registerJSON))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), http.StatusCreated, w.Code)

	// Attempt to register with the same email
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(registerJSON))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), http.StatusInternalServerError, w.Code) // This should return an error
}

func (s *AuthIntegrationTestSuite) TestLogin_InvalidCredentials() {
	// First register a user
	registerBody := map[string]interface{}{
		"email":      "invalid@example.com",
		"password":   "password123",
		"first_name": "Invalid",
		"last_name":  "Test",
	}
	registerJSON, _ := json.Marshal(registerBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(registerJSON))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), http.StatusCreated, w.Code)

	// Try to login with wrong password
	loginBody := map[string]interface{}{
		"email":    "invalid@example.com",
		"password": "wrongpassword",
	}
	loginJSON, _ := json.Marshal(loginBody)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *AuthIntegrationTestSuite) TestProfile_Unauthorized() {
	// Try to access profile without token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/profile", nil)
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)

	// Try to access profile with invalid token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/profile", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

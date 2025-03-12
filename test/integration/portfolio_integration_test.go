// test/integration/portfolio_integration_test.go
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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PortfolioIntegrationTestSuite struct {
	suite.Suite
	router           *gin.Engine
	userRepo         repository.UserRepository
	portfolioRepo    repository.PortfolioRepository
	userService      services.UserService
	portfolioService services.PortfolioService
	tokenService     auth.TokenService
	cfg              *config.Config
	authToken        string
	userID           uuid.UUID
}

func (s *PortfolioIntegrationTestSuite) SetupSuite() {
	// Use test database
	db := GetTestDB()
	s.userRepo = repository.NewUserRepository(db)
	s.portfolioRepo = repository.NewPortfolioRepository(db)
	s.userService = services.NewUserService(s.userRepo)
	s.portfolioService = services.NewPortfolioService(s.portfolioRepo)

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

	// Create handlers
	authHandler := handlers.NewAuthHandler(s.userService, s.tokenService)
	portfolioHandler := handlers.NewPortfolioHandler(s.portfolioService)

	// Set up auth routes
	authGroup := s.router.Group("/api/v1/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
	}

	// Protected routes (with auth middleware)
	protected := s.router.Group("/api/v1")
	protected.Use(auth.AuthMiddleware(s.tokenService))
	{
		protected.POST("/portfolio", portfolioHandler.CreatePortfolio)
		protected.GET("/portfolio", portfolioHandler.GetPortfolio)
		protected.GET("/portfolio/holdings", portfolioHandler.GetHoldings)
		protected.POST("/portfolio/holdings", portfolioHandler.AddHolding)
		protected.DELETE("/portfolio/holdings/:symbol", portfolioHandler.RemoveHolding)
	}

	// Create a test user and get auth token
	s.createTestUser()
}

func (s *PortfolioIntegrationTestSuite) createTestUser() {
	// Register a test user
	registerBody := map[string]interface{}{
		"email":      "portfolio_test@example.com",
		"password":   "password123",
		"first_name": "Portfolio",
		"last_name":  "Test",
	}
	registerJSON, _ := json.Marshal(registerBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(registerJSON))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil || w.Code != http.StatusCreated {
		s.T().Fatalf("Failed to create test user: %v", err)
	}

	s.authToken = response["token"].(string)
	user := response["user"].(map[string]interface{})
	userID, err := uuid.Parse(user["id"].(string))
	if err != nil {
		s.T().Fatalf("Failed to parse user ID: %v", err)
	}
	s.userID = userID
}

func TestPortfolioIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PortfolioIntegrationTestSuite))
}

func (s *PortfolioIntegrationTestSuite) TestCreateAndGetPortfolio() {
	// Create a portfolio
	createBody := map[string]interface{}{
		"initial_balance": 10000.0,
	}
	createJSON, _ := json.Marshal(createBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/portfolio", bytes.NewBuffer(createJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	s.router.ServeHTTP(w, req)

	// Check if we got a 201 (created) or 409 (conflict - already exists)
	if w.Code == http.StatusCreated {
		// New portfolio created - validate the response
		var createResponse map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &createResponse)
		assert.NoError(s.T(), err)
		assert.NotNil(s.T(), createResponse["portfolio"])
	} else if w.Code == http.StatusConflict {
		// Portfolio already exists - that's okay too
		assert.Equal(s.T(), http.StatusConflict, w.Code)
	} else {
		// Any other status is an error
		assert.Fail(s.T(), "Expected status 201 or 409, got %d", w.Code)
	}

	// Get the portfolio
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/portfolio", nil)
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	s.router.ServeHTTP(w, req)

	var getResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &getResponse)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, w.Code)
	assert.NotNil(s.T(), getResponse["portfolio"])

	retrievedPortfolio := getResponse["portfolio"].(map[string]interface{})
	assert.Equal(s.T(), float64(10000.0), retrievedPortfolio["total_value"])
}

func (s *PortfolioIntegrationTestSuite) TestAddAndGetHoldings() {
	// First ensure portfolio exists
	s.TestCreateAndGetPortfolio()

	// Add a holding
	addBody := map[string]interface{}{
		"symbol":   "AAPL",
		"quantity": 10.0,
		"price":    150.0,
	}
	addJSON, _ := json.Marshal(addBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/portfolio/holdings", bytes.NewBuffer(addJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	// Get holdings
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/portfolio/holdings", nil)
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	s.router.ServeHTTP(w, req)

	var getResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &getResponse)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, w.Code)
	assert.NotNil(s.T(), getResponse["holdings"])

	holdings := getResponse["holdings"].([]interface{})
	assert.GreaterOrEqual(s.T(), len(holdings), 1)

	// Verify holding details
	found := false
	for _, h := range holdings {
		holding := h.(map[string]interface{})
		if holding["symbol"] == "AAPL" {
			found = true
			assert.Equal(s.T(), float64(10.0), holding["quantity"])
			assert.Equal(s.T(), float64(150.0), holding["average_cost"])
			assert.Equal(s.T(), float64(150.0), holding["current_price"])
			assert.Equal(s.T(), float64(1500.0), holding["market_value"]) // 10 * 150
			break
		}
	}
	assert.True(s.T(), found, "AAPL holding not found")

	// Remove a holding
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/portfolio/holdings/AAPL", nil)
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	// Get holdings again to verify removal
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/portfolio/holdings", nil)
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	s.router.ServeHTTP(w, req)

	err = json.Unmarshal(w.Body.Bytes(), &getResponse)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, w.Code)

	holdings = getResponse["holdings"].([]interface{})
	found = false
	for _, h := range holdings {
		holding := h.(map[string]interface{})
		if holding["symbol"] == "AAPL" {
			found = true
			break
		}
	}
	assert.False(s.T(), found, "AAPL holding should be removed")
}

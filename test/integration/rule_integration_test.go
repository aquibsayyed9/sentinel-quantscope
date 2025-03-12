// test/integration/rule_integration_test.go
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

type RuleIntegrationTestSuite struct {
	suite.Suite
	router       *gin.Engine
	userRepo     repository.UserRepository
	ruleRepo     repository.RuleRepository
	userService  services.UserService
	ruleService  services.RuleService
	tokenService auth.TokenService
	cfg          *config.Config
	authToken    string
	userID       uuid.UUID
}

func (s *RuleIntegrationTestSuite) SetupSuite() {
	// Use test database
	db := GetTestDB()
	s.userRepo = repository.NewUserRepository(db)
	s.ruleRepo = repository.NewRuleRepository(db)
	s.userService = services.NewUserService(s.userRepo)
	s.ruleService = services.NewRuleService(s.ruleRepo)

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
	ruleHandler := handlers.NewRuleHandler(s.ruleService)

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
		protected.POST("/rules", ruleHandler.CreateRule)
		protected.GET("/rules", ruleHandler.GetRules)
		protected.GET("/rules/:id", ruleHandler.GetRule)
		protected.PUT("/rules/:id/activate", ruleHandler.ActivateRule)
		protected.PUT("/rules/:id/deactivate", ruleHandler.DeactivateRule)
		protected.DELETE("/rules/:id", ruleHandler.DeleteRule)
	}

	// Create a test user and get auth token
	s.createTestUser()
}

func (s *RuleIntegrationTestSuite) createTestUser() {
	// Register a test user
	registerBody := map[string]interface{}{
		"email":      "rule_test@example.com",
		"password":   "password123",
		"first_name": "Rule",
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

func TestRuleIntegrationSuite(t *testing.T) {
	suite.Run(t, new(RuleIntegrationTestSuite))
}

func (s *RuleIntegrationTestSuite) TestCreateAndGetRule() {
	// Create a new rule
	createBody := map[string]interface{}{
		"name":        "Test Rule",
		"description": "A test trading rule",
		"symbol":      "AAPL",
		"rule_type":   "stop_loss",
		"conditions": []map[string]interface{}{
			{
				"type":     "price",
				"symbol":   "AAPL",
				"operator": "less_than",
				"value":    150.0,
			},
		},
		"actions": []map[string]interface{}{
			{
				"type":       "sell",
				"symbol":     "AAPL",
				"quantity":   10.0,
				"order_type": "market",
			},
		},
	}
	createJSON, _ := json.Marshal(createBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/rules", bytes.NewBuffer(createJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	s.router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResponse)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, w.Code)
	assert.NotNil(s.T(), createResponse["rule"])

	rule := createResponse["rule"].(map[string]interface{})
	ruleID := rule["id"].(string)
	assert.Equal(s.T(), "Test Rule", rule["name"])
	assert.Equal(s.T(), "A test trading rule", rule["description"])
	assert.Equal(s.T(), "AAPL", rule["symbol"])
	assert.Equal(s.T(), "stop_loss", rule["rule_type"])
	assert.Equal(s.T(), "active", rule["status"])

	// Get the rule by ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/rules/"+ruleID, nil)
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	s.router.ServeHTTP(w, req)

	var getResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &getResponse)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, w.Code)
	assert.NotNil(s.T(), getResponse["rule"])

	retrievedRule := getResponse["rule"].(map[string]interface{})
	assert.Equal(s.T(), ruleID, retrievedRule["id"])
	assert.Equal(s.T(), "Test Rule", retrievedRule["name"])

	// Get all rules
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/rules", nil)
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	s.router.ServeHTTP(w, req)

	var listResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &listResponse)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, w.Code)
	assert.NotNil(s.T(), listResponse["rules"])

	rules := listResponse["rules"].([]interface{})
	assert.GreaterOrEqual(s.T(), len(rules), 1)
}

func (s *RuleIntegrationTestSuite) TestActivateDeactivateRule() {
	// First create a rule
	createBody := map[string]interface{}{
		"name":        "Activation Test Rule",
		"description": "Testing activation/deactivation",
		"symbol":      "TSLA",
		"rule_type":   "take_profit",
		"conditions": []map[string]interface{}{
			{
				"type":     "price",
				"symbol":   "TSLA",
				"operator": "greater_than",
				"value":    900.0,
			},
		},
		"actions": []map[string]interface{}{
			{
				"type":       "sell",
				"symbol":     "TSLA",
				"quantity":   5.0,
				"order_type": "market",
			},
		},
	}
	createJSON, _ := json.Marshal(createBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/rules", bytes.NewBuffer(createJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	s.router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResponse)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, w.Code)

	rule := createResponse["rule"].(map[string]interface{})
	ruleID := rule["id"].(string)
	assert.Equal(s.T(), "active", rule["status"])

	// Deactivate the rule
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/v1/rules/"+ruleID+"/deactivate", nil)
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	// Get the rule to check its status
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/rules/"+ruleID, nil)
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	s.router.ServeHTTP(w, req)

	var getResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &getResponse)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, w.Code)

	retrievedRule := getResponse["rule"].(map[string]interface{})
	assert.Equal(s.T(), "inactive", retrievedRule["status"])

	// Reactivate the rule
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/v1/rules/"+ruleID+"/activate", nil)
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	// Get the rule again to check its status
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/rules/"+ruleID, nil)
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	s.router.ServeHTTP(w, req)

	err = json.Unmarshal(w.Body.Bytes(), &getResponse)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, w.Code)

	retrievedRule = getResponse["rule"].(map[string]interface{})
	assert.Equal(s.T(), "active", retrievedRule["status"])
}

func (s *RuleIntegrationTestSuite) TestDeleteRule() {
	// First create a rule
	createBody := map[string]interface{}{
		"name":        "Deletion Test Rule",
		"description": "Testing rule deletion",
		"symbol":      "MSFT",
		"rule_type":   "buy",
		"conditions": []map[string]interface{}{
			{
				"type":     "price",
				"symbol":   "MSFT",
				"operator": "less_than",
				"value":    300.0,
			},
		},
		"actions": []map[string]interface{}{
			{
				"type":       "buy",
				"symbol":     "MSFT",
				"quantity":   2.0,
				"order_type": "market",
			},
		},
	}
	createJSON, _ := json.Marshal(createBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/rules", bytes.NewBuffer(createJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	s.router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResponse)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, w.Code)

	rule := createResponse["rule"].(map[string]interface{})
	ruleID := rule["id"].(string)

	// Delete the rule
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/rules/"+ruleID, nil)
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	// Try to get the deleted rule
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/rules/"+ruleID, nil)
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code) // Should return error as rule is deleted
}

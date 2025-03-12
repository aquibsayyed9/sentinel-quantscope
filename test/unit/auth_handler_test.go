// test/unit/auth_handler_test.go
package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aquibsayyed9/sentinel/internal/auth"
	"github.com/aquibsayyed9/sentinel/internal/handlers"
	"github.com/aquibsayyed9/sentinel/internal/models"
	"github.com/aquibsayyed9/sentinel/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mock user service
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, email, password, firstName, lastName string) (*models.User, error) {
	args := m.Called(ctx, email, password, firstName, lastName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) Login(ctx context.Context, email, password string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// Mock token service
type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateToken(userID uuid.UUID, email string) (string, error) {
	args := m.Called(userID, email)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) ValidateToken(tokenString string) (*auth.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

type AuthHandlerTestSuite struct {
	suite.Suite
	router      *gin.Engine
	mockService *MockUserService
	mockToken   *MockTokenService
	authHandler *handlers.AuthHandler
}

func (s *AuthHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.router = gin.Default()
	s.mockService = new(MockUserService)
	s.mockToken = new(MockTokenService)
	s.authHandler = handlers.NewAuthHandler(s.mockService, s.mockToken)

	// Set up routes
	auth := s.router.Group("/api/v1/auth")
	{
		auth.POST("/register", s.authHandler.Register)
		auth.POST("/login", s.authHandler.Login)
	}

	// Protected routes (simulate auth middleware)
	protected := s.router.Group("/api/v1")
	protected.Use(func(c *gin.Context) {
		c.Set("userID", uuid.New())
		c.Next()
	})
	{
		protected.GET("/profile", s.authHandler.GetProfile)
	}
}

func TestAuthHandlerSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

func (s *AuthHandlerTestSuite) TestRegister_Success() {
	// Arrange
	reqBody := map[string]interface{}{
		"email":      "test@example.com",
		"password":   "password123",
		"first_name": "John",
		"last_name":  "Doe",
	}
	reqJSON, _ := json.Marshal(reqBody)

	userId := uuid.New()
	mockUser := &models.User{
		ID:          userId,
		Email:       "test@example.com",
		FirstName:   "John",
		LastName:    "Doe",
		AccountType: "free",
	}

	mockToken := "mock.jwt.token"

	s.mockService.On("Register", mock.Anything, "test@example.com", "password123", "John", "Doe").Return(mockUser, nil)
	s.mockToken.On("GenerateToken", userId, "test@example.com").Return(mockToken, nil)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	// Assert
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, w.Code)
	assert.NotNil(s.T(), response["user"])
	assert.Equal(s.T(), mockToken, response["token"])

	user := response["user"].(map[string]interface{})
	assert.Equal(s.T(), userId.String(), user["id"])
	assert.Equal(s.T(), "test@example.com", user["email"])
	assert.Equal(s.T(), "John", user["first_name"])
	assert.Equal(s.T(), "Doe", user["last_name"])
	assert.Equal(s.T(), "free", user["account_type"])

	s.mockService.AssertExpectations(s.T())
	s.mockToken.AssertExpectations(s.T())
}

func (s *AuthHandlerTestSuite) TestRegister_ValidationError() {
	// Arrange - missing required fields
	reqBody := map[string]interface{}{
		"email": "test@example.com",
		// Missing password, first_name, last_name
	}
	reqJSON, _ := json.Marshal(reqBody)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	s.mockService.AssertNotCalled(s.T(), "Register")
	s.mockToken.AssertNotCalled(s.T(), "GenerateToken")
}

func (s *AuthHandlerTestSuite) TestRegister_ServiceError() {
	// Arrange
	reqBody := map[string]interface{}{
		"email":      "test@example.com",
		"password":   "password123",
		"first_name": "John",
		"last_name":  "Doe",
	}
	reqJSON, _ := json.Marshal(reqBody)

	expectedErr := errors.New("service error")
	s.mockService.On("Register", mock.Anything, "test@example.com", "password123", "John", "Doe").Return(nil, expectedErr)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)

	s.mockService.AssertExpectations(s.T())
	s.mockToken.AssertNotCalled(s.T(), "GenerateToken")
}

func (s *AuthHandlerTestSuite) TestLogin_Success() {
	// Arrange
	reqBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	}
	reqJSON, _ := json.Marshal(reqBody)

	userId := uuid.New()
	mockUser := &models.User{
		ID:          userId,
		Email:       "test@example.com",
		FirstName:   "John",
		LastName:    "Doe",
		AccountType: "free",
	}

	mockToken := "mock.jwt.token"

	s.mockService.On("Login", mock.Anything, "test@example.com", "password123").Return(mockUser, nil)
	s.mockToken.On("GenerateToken", userId, "test@example.com").Return(mockToken, nil)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	// Assert
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, w.Code)
	assert.NotNil(s.T(), response["user"])
	assert.Equal(s.T(), mockToken, response["token"])

	user := response["user"].(map[string]interface{})
	assert.Equal(s.T(), userId.String(), user["id"])
	assert.Equal(s.T(), "test@example.com", user["email"])

	s.mockService.AssertExpectations(s.T())
	s.mockToken.AssertExpectations(s.T())
}

func (s *AuthHandlerTestSuite) TestLogin_InvalidCredentials() {
	// Arrange
	reqBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	}
	reqJSON, _ := json.Marshal(reqBody)

	s.mockService.On("Login", mock.Anything, "test@example.com", "password123").Return(nil, services.ErrInvalidCredentials)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)

	s.mockService.AssertExpectations(s.T())
	s.mockToken.AssertNotCalled(s.T(), "GenerateToken")
}

func (s *AuthHandlerTestSuite) TestGetProfile_Success() {
	// Arrange
	userId := uuid.New()
	mockUser := &models.User{
		ID:          userId,
		Email:       "test@example.com",
		FirstName:   "John",
		LastName:    "Doe",
		AccountType: "free",
	}

	s.mockService.On("GetUserByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(mockUser, nil)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/profile", nil)
	s.router.ServeHTTP(w, req)

	// Assert
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, w.Code)
	assert.NotNil(s.T(), response["user"])

	user := response["user"].(map[string]interface{})
	assert.Equal(s.T(), userId.String(), user["id"])
	assert.Equal(s.T(), "test@example.com", user["email"])
	assert.Equal(s.T(), "John", user["first_name"])
	assert.Equal(s.T(), "Doe", user["last_name"])
	assert.Equal(s.T(), "free", user["account_type"])

	s.mockService.AssertExpectations(s.T())
}

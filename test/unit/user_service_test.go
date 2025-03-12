// test/unit/user_service_test.go
package unit

import (
	"context"
	"errors"
	"testing"

	"github.com/aquibsayyed9/sentinel/internal/models"
	"github.com/aquibsayyed9/sentinel/internal/repository"
	"github.com/aquibsayyed9/sentinel/internal/services"
	"github.com/aquibsayyed9/sentinel/test/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type UserServiceTestSuite struct {
	suite.Suite
	mockRepo *mocks.MockUserRepository
	service  services.UserService
}

func (s *UserServiceTestSuite) SetupTest() {
	s.mockRepo = new(mocks.MockUserRepository)
	s.service = services.NewUserService(s.mockRepo)
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (s *UserServiceTestSuite) TestRegister_Success() {
	// Arrange
	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	firstName := "John"
	lastName := "Doe"

	// The mock should expect the Create method to be called with a user object
	s.mockRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	// Act
	user, err := s.service.Register(ctx, email, password, firstName, lastName)

	// Assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), user)
	assert.Equal(s.T(), email, user.Email)
	assert.Equal(s.T(), firstName, user.FirstName)
	assert.Equal(s.T(), lastName, user.LastName)
	assert.Equal(s.T(), "free", user.AccountType)

	// Verify that the password was hashed
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	assert.NoError(s.T(), err)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestRegister_RepositoryError() {
	// Arrange
	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	firstName := "John"
	lastName := "Doe"
	expectedErr := errors.New("database error")

	s.mockRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(expectedErr)

	// Act
	user, err := s.service.Register(ctx, email, password, firstName, lastName)

	// Assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), expectedErr, err)
	assert.Nil(s.T(), user)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestLogin_Success() {
	// Arrange
	ctx := context.Background()
	email := "test@example.com"
	password := "password123"

	// Hash the password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mockUser := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    "John",
		LastName:     "Doe",
		AccountType:  "free",
	}

	s.mockRepo.On("GetByEmail", ctx, email).Return(mockUser, nil)
	s.mockRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	// Act
	user, err := s.service.Login(ctx, email, password)

	// Assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), user)
	assert.Equal(s.T(), mockUser.ID, user.ID)
	assert.Equal(s.T(), email, user.Email)
	assert.NotNil(s.T(), user.LastLogin)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestLogin_InvalidCredentials() {
	// Arrange
	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	wrongPassword := "wrongpassword"

	// Hash the password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mockUser := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    "John",
		LastName:     "Doe",
		AccountType:  "free",
	}

	s.mockRepo.On("GetByEmail", ctx, email).Return(mockUser, nil)

	// Act
	user, err := s.service.Login(ctx, email, wrongPassword)

	// Assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), services.ErrInvalidCredentials, err)
	assert.Nil(s.T(), user)

	s.mockRepo.AssertExpectations(s.T())
}

func (s *UserServiceTestSuite) TestLogin_UserNotFound() {
	// Arrange
	ctx := context.Background()
	email := "nonexistent@example.com"
	password := "password123"

	s.mockRepo.On("GetByEmail", ctx, email).Return(nil, repository.ErrUserNotFound)

	// Act
	user, err := s.service.Login(ctx, email, password)

	// Assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), services.ErrInvalidCredentials, err)
	assert.Nil(s.T(), user)

	s.mockRepo.AssertExpectations(s.T())
}

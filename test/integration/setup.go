// test/integration/setup.go
package integration

import (
	"fmt"
	"log"
	"os"
	"testing"

	// "github.com/aquibsayyed9/sentinel/internal/config"
	// "github.com/aquibsayyed9/sentinel/internal/db"
	"github.com/aquibsayyed9/sentinel/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	testDB *gorm.DB
)

// SetupIntegrationTest initializes a test database for integration tests
func SetupIntegrationTest() (*gorm.DB, error) {
	// Use a real database connection string for testing
	dsn := "host=localhost user=postgres password=postgres dbname=sentinel_test port=5432 sslmode=disable"

	// Connect to the database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	// Migrate the database schema for testing
	err = db.AutoMigrate(
		&models.User{},
		&models.TradingRule{},
		&models.Execution{},
		&models.Portfolio{},
		&models.PortfolioHolding{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate test database: %w", err)
	}

	// Clean the database before tests
	if err := cleanTestDatabase(db); err != nil {
		return nil, fmt.Errorf("failed to clean test database: %w", err)
	}

	return db, nil
}

// cleanTestDatabase truncates all tables in the test database
func cleanTestDatabase(db *gorm.DB) error {
	// Use raw SQL to disable foreign key checks
	if err := db.Exec("SET session_replication_role = 'replica';").Error; err != nil {
		return err
	}

	// Truncate tables
	if err := db.Exec("TRUNCATE users, trading_rules, executions, portfolios, portfolio_holdings RESTART IDENTITY CASCADE;").Error; err != nil {
		return err
	}

	// Re-enable foreign key checks
	return db.Exec("SET session_replication_role = 'origin';").Error
}

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	var err error
	testDB, err = SetupIntegrationTest()
	if err != nil {
		log.Fatalf("Failed to set up integration test: %v", err)
	}

	// Run the tests
	code := m.Run()

	// Clean up test database
	if err := cleanTestDatabase(testDB); err != nil {
		log.Printf("Failed to clean test database: %v", err)
	}

	os.Exit(code)
}

// GetTestDB returns the test database instance
func GetTestDB() *gorm.DB {
	if testDB == nil {
		log.Println("WARNING: testDB is nil - database might not be initialized")
		// You could initialize it here as a fallback
		var err error
		testDB, err = SetupIntegrationTest()
		if err != nil {
			log.Printf("Failed to initialize database on demand: %v", err)
			return nil
		}
	}
	return testDB
}

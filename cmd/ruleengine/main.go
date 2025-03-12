// cmd/ruleengine/main.go
package main

import (
	"context"
	"log"
	"time"

	"github.com/aquibsayyed9/sentinel/internal/config"
	"github.com/aquibsayyed9/sentinel/internal/db"
	"github.com/aquibsayyed9/sentinel/internal/models"
	"github.com/aquibsayyed9/sentinel/internal/repository"
	"github.com/aquibsayyed9/sentinel/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	l := log.New(log.Writer(), "[RULE ENGINE] ", log.LstdFlags)
	l.Println("Initializing rule engine...")

	// Connect to database
	database, err := db.ConnectDatabase(cfg)
	if err != nil {
		l.Fatalf("Failed to connect to database: %v", err)
	}
	l.Println("Database connection established")

	// Initialize repositories
	ruleRepo := repository.NewRuleRepository(database)
	marketDataRepo := repository.NewMarketDataRepository(database)
	portfolioRepo := repository.NewPortfolioRepository(database)
	executionRepo := repository.NewExecutionRepository(database)

	// Initialize services
	marketDataService := services.NewMarketDataService(marketDataRepo)
	portfolioService := services.NewPortfolioService(portfolioRepo)
	executionService := services.NewExecutionService(executionRepo, ruleRepo)

	// Create rule engine service
	ruleEngineService := services.NewRuleEngineService(
		ruleRepo,
		marketDataService,
		portfolioService,
		executionService,
	)

	l.Println("Services initialized")
	l.Println("Starting rule evaluation loop...")

	// Rule evaluation loop
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.Println("Evaluating trading rules...")

			// Get all active rules
			ctx := context.Background()
			rules, err := ruleRepo.GetActiveRules(ctx)
			if err != nil {
				l.Printf("Failed to get active rules: %v", err)
				continue
			}

			l.Printf("Found %d active rules to evaluate", len(rules))

			// Evaluate each rule
			for _, rule := range rules {
				l.Printf("Evaluating rule %s: %s", rule.ID, rule.Name)

				// Check if rule should be executed
				shouldExecute, err := ruleEngineService.EvaluateRule(ctx, &rule)
				if err != nil {
					l.Printf("Error evaluating rule %s: %v", rule.ID, err)
					continue
				}

				if shouldExecute {
					l.Printf("Rule %s triggered! Executing...", rule.ID)

					// Get current market data
					marketData, err := marketDataService.GetPrice(ctx, rule.Symbol)
					if err != nil {
						l.Printf("Failed to get price for %s: %v", rule.Symbol, err)
						continue
					}

					// Execute the rule
					if err := ruleEngineService.ExecuteRule(ctx, &rule, marketData.Close); err != nil {
						l.Printf("Error executing rule %s: %v", rule.ID, err)
					} else {
						l.Printf("Rule %s executed successfully", rule.ID)

						// Create execution record
						execution := &models.RuleExecution{
							RuleID:       rule.ID,
							UserID:       rule.UserID,
							Symbol:       rule.Symbol,
							ExecutedAt:   time.Now(),
							TriggerPrice: marketData.Close,
							Status:       "executed",
							Notes:        "Automated execution by rule engine",
						}

						if err := executionService.CreateExecution(ctx, execution); err != nil {
							l.Printf("Failed to record execution for rule %s: %v", rule.ID, err)
						}
					}
				} else {
					l.Printf("Rule %s conditions not met, skipping execution", rule.ID)
				}
			}

			l.Println("Rule evaluation cycle completed")
		}
	}
}

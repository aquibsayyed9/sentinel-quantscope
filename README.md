# Sentinel

Sentinel is an automated trading platform that allows users to create rule-based trading strategies and execute them automatically.

## Features

- User authentication and management
- Trading rule creation with custom conditions
- Real-time market data streaming
- Automated trade execution
- Portfolio tracking and performance analytics
- AI-driven trading insights

## Getting Started

### Prerequisites

- Go 1.20 or higher
- PostgreSQL 14 or higher
- Docker and Docker Compose (for local development)

### Local Development

1. Clone the repository:
git clone https://github.com/aquibsayyed9/sentinel.git
cd sentinel

2. Set up the environment:
cp configs/app.yaml configs/app.local.yaml

3. Start the database:
docker-compose up -d postgres


4. Run database migrations:
make migrate

5. Run the API server:
make run-api

6. Access the API at `http://localhost:8080`

## Project Structure

- `cmd/`: Application entry points
- `internal/`: Private application code
- `pkg/`: Shareable packages
- `migrations/`: Database migrations
- `configs/`: Configuration files
- `docs/`: Documentation
- `test/`: Test files

## License



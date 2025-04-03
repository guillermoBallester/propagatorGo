# PropagatorGo

PropagatorGo is a Go-based system for scraping, processing, and storing financial news articles related to specific stock symbols. The application uses a worker pool architecture to efficiently collect news from various sources (currently Yahoo Finance), store them in a PostgreSQL database, and make them available through a RESTful API.

## Overview

PropagatorGo consists of several key components:

1. **Web Scrapers**: Collect financial news articles from configured sources
2. **Worker Pools**: Manage concurrent processing of multiple stock symbols
3. **Task System**: Handle work distribution and queuing
4. **Scheduler**: Coordinate when scraping and processing jobs run
5. **Orchestrator**: Manage the lifecycle of worker pools
6. **Database**: Store collected articles for later retrieval
7. **API**: Provide access to collected data

## Project Structure

```
stockalpha/
├── cmd/
│   └── propagator/           # Main application entry point
├── internal/
│   ├── api/                  # RESTful API components
│   │   ├── handlers/         # Request handlers
│   │   ├── middleware/       # HTTP middleware
│   │   ├── response/         # Standardized response formats
│   │   ├── router/           # URL routing
│   │   └── server.go         # API server setup
│   ├── config/               # Configuration structures and loading
│   ├── constants/            # Application-wide constants
│   ├── database/             # Database connectivity and models
│   │   ├── migrations/       # SQL schema definitions
│   │   ├── queries/          # SQL query definitions
│   │   └── sqlc/             # Generated database code (sqlc)
│   ├── model/                # Domain models
│   ├── orchestrator/         # Worker pool orchestration
│   ├── queue/                # Message queue implementation (Redis)
│   ├── repository/           # Data access layer
│   ├── scheduler/            # Job scheduling
│   ├── scrapper/             # Web scraping implementation
│   ├── task/                 # Task definition and processing
│   └── worker/               # Worker pool implementation
├── .gitignore
├── .golangci.yml             # Linter configuration
├── config.json               # Application configuration
├── docker-compose.yml        # Container orchestration
├── Dockerfile                # Container definition
├── Makefile                  # Build and development commands
└── sqlc.yaml                 # SQL code generation config
```

## Core Components Explained

### Scheduler + Orchestrator + Task System

The combination of the Scheduler, Orchestrator, and Task system forms the backbone of the application's asynchronous processing capabilities.

#### Scheduler (`internal/scheduler/scheduler.go`)

The Scheduler is responsible for:

- **Time-based Execution**: Running jobs at specified intervals using cron expressions
- **Job Management**: Tracking job status, last run time, and execution results
- **Timeout Handling**: Ensuring jobs don't run indefinitely

#### Orchestrator (`internal/orchestrator/orchestrator.go`)

The Orchestrator manages the lifecycle of worker pools:

- **Worker Pool Creation**: Initializing pools of workers based on configuration
- **Job Registration**: Connecting scheduled jobs to worker pools
- **Coordination**: Starting and stopping pools in response to scheduled events
- **Resource Management**: Controlling the number of concurrent workers

#### Task System (`internal/task/task.go` and `internal/task/service.go`)

The Task system provides:

- **Work Distribution**: Defining units of work that can be processed independently
- **Queueing**: Using Redis to store and retrieve tasks asynchronously
- **Progress Tracking**: Monitoring task completion status
- **Parameter Passing**: Transferring data between components in a structured way

### Problems Solved by This Architecture

1. **Scalability**: The worker pool model allows easy scaling by adjusting the number of workers
2. **Fault Tolerance**: Failed tasks don't affect the entire system
3. **Resource Management**: Controlled concurrency prevents overwhelming external systems
4. **Decoupling**: Producers (scrapers) and consumers (database writers) operate independently
5. **Scheduling**: Time-based operations run automatically without manual intervention
6. **Flexibility**: New data sources can be added by creating new worker configurations

## Worker Types

StockAlpha implements two main worker types:

1. **Scraper Workers**: Collect news articles from configured sources
2. **Consumer Workers**: Process collected articles and store them in the database

## Data Flow

1. The Scheduler triggers a scraping job according to its cron schedule
2. The Orchestrator starts a pool of Scraper workers
3. Each worker processes stock symbols from the configured list
4. Articles are collected and published as tasks to a Redis queue
5. Consumer workers retrieve tasks from the queue and store articles in PostgreSQL
6. The API server provides endpoints to access the stored articles

## Configuration

Configuration is managed through a `config.json` file with sections for:

- **App**: General application settings
- **Scraper**: Web scraping configuration
- **Scheduler**: Job scheduling settings
- **Redis**: Message queue connection details
- **Database**: PostgreSQL connection parameters
- **StockList**: List of stock symbols to track

## API Endpoints

The application provides the following API endpoints:

- `GET /propagatorGo/v1/stocks/{symbol}/news`: Retrieves news for a specific stock symbol
- `GET /propagatorGo/v1/sources/{site}/news`: Retrieves news from a specific source
- `GET /propagatorGo/v1/health`: Returns the health status of the API

## Running the Application

### Prerequisites

- Docker and Docker Compose
- Go 1.23 or later (for development)

### Using Docker Compose

```bash
docker-compose up -d
```

### Development

```bash
# Build the application
make build

# Run the application
make run

# Run tests
make test

# Run linter
make lint
```

## Future Improvements

- Add more news sources
- Implement sentiment analysis for articles
- Create a web UI for viewing collected data
- Add user authentication to the API
- Implement rate limiting for external API calls

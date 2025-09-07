# Vision Data API

A REST API developed in Go using the Gin framework, integrated with Elasticsearch for logging, Redis for caching, and MongoDB for data persistence.

## 📁 Project Structure

```
├── cmd/                          # Application entry point
│   └── api/
│       ├── main.go              # Main application file
│       └── main_test.go         # Main tests
├── docker-compose.yml           # Docker Compose configuration
├── dockerfile                   # Application Docker image
├── go.mod                       # Go dependencies
├── go.sum                       # Dependencies checksums
├── index_tickets.json          # Elasticsearch index configuration
├── internal/                    # Internal application code
│   ├── config/
│   │   └── config.go           # Application configuration
│   ├── middleware/             # HTTP middlewares
│   │   ├── cors.go            # CORS middleware
│   │   ├── id.go              # ID generation middleware
│   │   ├── jwt.go             # JWT authentication middleware
│   │   ├── logger.go          # Logging middleware
│   │   ├── server.go          # Server configuration
│   │   └── throttling.go      # Rate limiting middleware
│   ├── models/
│   │   └── generic_response.go # Generic response models
│   ├── repositories/           # Data access layer
│   │   ├── elsearch/
│   │   │   └── connection.go   # Elasticsearch connection
│   │   ├── mongo/
│   │   │   └── connection.go   # MongoDB connection
│   │   └── redis/
│   │       ├── connection.go   # Redis connection
│   │       └── methods.go      # Redis methods
│   ├── routes/
│   │   └── routes.go          # API routes definition
│   ├── service/               # Service layer
│   │   └── healthcheck/
│   │       └── health.go      # Health check service
│   └── utils/
│       └── hosts.go           # Host utilities
├── pkg/                       # Reusable packages
│   └── logger/
│       └── logger.go         # Custom logging system
└── README.md                 # This file
```

## 🏗️ Architecture

The application follows a layered architecture:

- **cmd/**: Entry point and initial configuration
- **internal/**: Application-specific code (non-exportable)
  - **config/**: Configuration management
  - **middleware/**: HTTP middlewares (CORS, JWT, logging, etc.)
  - **models/**: Data structures
  - **repositories/**: Data access (Elasticsearch, MongoDB, Redis)
  - **routes/**: API routes definition
  - **service/**: Business logic
  - **utils/**: Various utilities
- **pkg/**: Reusable and exportable packages

## 🚀 How to Run

### Prerequisites

- Docker and Docker Compose installed
- SSL certificates in `./certs/` (for HTTPS)

### Using Docker Compose

Clone the repository and run:

```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f vision-data-api

# Stop services
docker-compose down

# Rebuild and start (after code changes)
docker-compose up -d --build
```

### Checking Service Status

```bash
# Check running containers
docker-compose ps

# Follow logs in real-time
docker-compose logs -f

# Access container shell
docker-compose exec vision-data-api sh
```

## ⚙️ Configuration

### Environment Variables

The application uses the following environment variables (configured in `.env`):

```bash
# Application settings - PRODUCTION
ENVIRONMENT_APP=prod
APP_PORT=8080
APP_PORT_TLS=8443
APP_CERT_FILE=/app/certs/server.crt
APP_KEY_FILE=/app/certs/server.key

# Elasticsearch - PRODUCTION
ELASTICSEARCH_URL=https://********:9200/
ELASTICSEARCH_USERNAME=elastic
ELASTICSEARCH_PASSWORD=**********

# Redis - PRODUCTION
REDIS_HOST=redis
REDIS_PORT=6379
```

### SSL Certificates

For HTTPS execution, place certificates in the `./certs/` folder:

- `server.crt` - Public certificate
- `server.key` - Private key

## 🔌 Endpoints

### Health Check

```
GET /healthcheck/
```

Returns the application status and its dependencies.

## 📊 Monitoring and Logs

### Elasticsearch

- URL: `https://********:9200/`
- Username: `elastic`
- Application logs are automatically sent to Elasticsearch

### Redis

- Host: `redis:6379`
- Used for caching and sessions

### Kibana

- Interface for Elasticsearch log visualization
- Configure access through Elasticsearch

## 🐳 Docker Services

The Docker Compose setup includes:

### vision-data-api

- **Ports**: 8080 (HTTP), 8443 (HTTPS)
- **Volumes**: SSL certificates mounted as read-only
- **Dependencies**: Redis

### redis

- **Image**: redis:7.4-alpine
- **Port**: 6379
- **Persistence**: Local volume for data storage
- **Configuration**: Appendonly disabled for performance

## 📝 Features

- **REST API**: Built with Gin Framework
- **Logging**: Elasticsearch integration via custom middleware
- **Caching**: Redis for performance optimization
- **Authentication**: JWT middleware for security
- **CORS**: Cross-origin access configuration
- **Rate Limiting**: Request throttling control
- **Health Check**: Application monitoring endpoint
- **HTTPS**: TLS support with custom certificates

## 🔧 Troubleshooting

### Common Issues

1. **Port conflicts**: Make sure ports 8080, 8443, and 6379 are available
2. **SSL certificates**: Ensure certificates are properly placed in `./certs/`
3. **Elasticsearch connection**: Verify the Elasticsearch URL and credentials
4. **Redis connection**: Check if Redis service is running

### Logs and Debugging

```bash
# View application logs
docker-compose logs vision-data-api

# View Redis logs
docker-compose logs redis

# View all services logs
docker-compose logs
```

## 🤝 Contributing

1. Fork the project
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License. See the `LICENSE` file for details.

## 🆘 Support

For support and questions:

- Open an issue in the repository
- Contact the development team

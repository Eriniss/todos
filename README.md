# Testbox - On-Premise & AWS Migration Testing Platform

A comprehensive learning project for testing on-premise services and AWS cloud migration strategies. This project implements a simple Todo List application with a focus on backend architecture, caching strategies, and message queuing.

## 🎯 Project Purpose

- **Educational**: Learn and experiment with various on-premise and AWS services
- **MVP**: Simple Todo CRUD operations with title and content
- **Migration Ready**: Designed for easy migration from on-premise to AWS services
- **Scalable**: Architecture supports load balancing, Kubernetes, and MongoDB sharding

## 🏗️ Architecture

### Current Stack (On-Premise)

**Backend:**
- Go 1.22 with Fiber v2 (HTTP framework)
- GORM (ORM)
- PostgreSQL (Primary database)
- Redis (Caching layer with Lazy Loading + Write-Through strategies)
- RabbitMQ (Message queue for async operations)
- MongoDB (Prepared for future replicaset + sharding tests)

**Frontend:**
- Vanilla HTML, CSS, JavaScript
- Nginx as web server and reverse proxy

**Infrastructure:**
- Docker & Docker Compose
- Ready for Kubernetes migration

### Future Migration Path (AWS)

- **PostgreSQL** → **DynamoDB**
- **RabbitMQ** → **SQS**
- **Fiber Backend** → **Elastic Beanstalk** or **ECS/EKS**
- **Load Balancing** → **ALB + Auto Scaling Groups**
- **Compute** → **EC2** or **ECS/EKS**

## 📁 Project Structure

```
testbox/
├── backend/                 # Go/Fiber backend
│   ├── cmd/
│   │   └── main.go         # Application entry point
│   ├── internal/
│   │   ├── api/            # HTTP handlers and routes
│   │   ├── cache/          # Redis caching implementation
│   │   ├── config/         # Configuration management
│   │   ├── database/       # Database connections
│   │   ├── messaging/      # RabbitMQ integration
│   │   ├── models/         # Data models
│   │   ├── repository/     # Data access layer
│   │   └── service/        # Business logic layer
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── frontend/               # Vanilla JS frontend
│   ├── css/
│   │   └── style.css
│   ├── js/
│   │   └── app.js
│   └── index.html
├── docs/                   # Documentation
├── docker-compose.yml      # Multi-container orchestration
├── nginx.conf             # Nginx configuration
├── Makefile               # Development shortcuts
└── README.md
```

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.22+ (for local development)
- Make (optional, for convenience)

### Running with Docker (Recommended)

1. **Clone and navigate to the project:**
   ```bash
   cd testbox
   ```

2. **Start all services:**
   ```bash
   make up
   # or
   docker-compose up -d
   ```

3. **Access the application:**
   - Frontend: http://localhost:8080
   - Backend API: http://localhost:3000/api/todos
   - Health Check: http://localhost:3000/health
   - RabbitMQ Management: http://localhost:15672 (guest/guest)

4. **View logs:**
   ```bash
   make logs
   # or
   docker-compose logs -f
   ```

5. **Stop services:**
   ```bash
   make down
   # or
   docker-compose down
   ```

### Local Development (Backend)

1. **Start infrastructure services:**
   ```bash
   docker-compose up -d postgres redis rabbitmq mongodb
   ```

2. **Run backend:**
   ```bash
   cd backend
   go mod download
   go run cmd/main.go
   ```

3. **Serve frontend:**
   ```bash
   cd frontend
   python3 -m http.server 8080
   ```

## 🔧 Configuration

Environment variables can be set in `.env` file (see `.env.example`):

```env
SERVER_PORT=3000
DB_HOST=localhost
DB_PORT=5432
REDIS_HOST=localhost
REDIS_PORT=6379
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
MONGO_URL=mongodb://localhost:27017
```

## 📚 Key Learning Points

### 1. Caching Strategies

**Lazy Loading (Cache-Aside):**
- Used in `GetTodo()` operation
- Check cache first → on miss, fetch from DB → populate cache
- Benefits: Only caches what's needed

**Write-Through:**
- Used in `CreateTodo()` and `UpdateTodo()` operations
- Write to DB first → immediately write to cache
- Benefits: Cache is always up-to-date

Implementation: [backend/internal/cache/redis.go](backend/internal/cache/redis.go)

### 2. Layered Architecture

```
Handler (API) → Service (Business Logic) → Repository (Data Access)
                    ↓                           ↓
                  Cache                      Database
                    ↓
                Messaging
```

### 3. Message Queue Integration

- Async event publishing for todo operations
- Prepared for migration to AWS SQS
- Implementation: [backend/internal/messaging/rabbitmq.go](backend/internal/messaging/rabbitmq.go)

### 4. Database Abstraction

- Repository pattern for easy database switching
- GORM for ORM operations
- Ready for DynamoDB migration

## 🔍 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/todos` | Get all todos |
| GET | `/api/todos/:id` | Get a specific todo |
| POST | `/api/todos` | Create a new todo |
| PUT | `/api/todos/:id` | Update a todo |
| DELETE | `/api/todos/:id` | Delete a todo |
| GET | `/health` | Health check |

### Example Requests

**Create Todo:**
```bash
curl -X POST http://localhost:3000/api/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Go","content":"Study Fiber v3 framework"}'
```

**Get All Todos:**
```bash
curl http://localhost:3000/api/todos
```

**Update Todo:**
```bash
curl -X PUT http://localhost:3000/api/todos/{id} \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Go","content":"Completed!","completed":true}'
```

## 🎓 Learning Roadmap

### Phase 1: On-Premise (Current)
- ✅ Go/Fiber backend with GORM
- ✅ PostgreSQL database
- ✅ Redis caching (Lazy Loading + Write-Through)
- ✅ RabbitMQ message queue
- ✅ Docker containerization
- 🔄 Kubernetes deployment (planned)
- 🔄 MongoDB replicaset + sharding (planned)

### Phase 2: AWS Migration
- 🔄 DynamoDB integration
- 🔄 SQS message queue
- 🔄 Elastic Beanstalk deployment
- 🔄 ALB + Auto Scaling Groups
- 🔄 ECS/EKS container orchestration

### Phase 3: Advanced Topics
- 🔄 Multi-region deployment
- 🔄 Blue-green deployment
- 🔄 Canary releases
- 🔄 Performance testing and optimization
- 🔄 Observability (logging, monitoring, tracing)

## 🛠️ Makefile Commands

```bash
make help           # Show all available commands
make up             # Start all services
make down           # Stop all services
make build          # Build all services
make logs           # Show logs
make clean          # Remove all containers and volumes
make restart        # Restart services
make health         # Check health of all services
make backend-dev    # Run backend in development mode
make frontend-dev   # Serve frontend locally
```

## 📊 Service Ports

| Service | Port | Description |
|---------|------|-------------|
| Frontend | 8080 | Nginx web server |
| Backend | 3000 | Go/Fiber API |
| PostgreSQL | 5432 | Database |
| Redis | 6379 | Cache |
| RabbitMQ | 5672 | Message queue |
| RabbitMQ UI | 15672 | Management interface |
| MongoDB | 27017 | Document database |

## 🧪 Testing

```bash
# Backend tests
cd backend
go test -v ./...

# Health check
make health
```

## 📝 Future Enhancements

- [ ] Add comprehensive unit and integration tests
- [ ] Implement MongoDB adapter alongside PostgreSQL
- [ ] Create Kubernetes manifests (deployments, services, ingress)
- [ ] Add CI/CD pipeline
- [ ] Implement AWS SDK for DynamoDB and SQS
- [ ] Add observability stack (Prometheus, Grafana)
- [ ] Implement authentication and authorization
- [ ] Add rate limiting and request throttling
- [ ] Create Terraform/CloudFormation templates for AWS

## 🤝 Contributing

This is a personal learning project, but suggestions and improvements are welcome!

## 📄 License

MIT License - Feel free to use this project for learning purposes.

---

**Note:** This project is designed for learning and experimentation. Security features like authentication, input validation, and rate limiting should be enhanced before any production use.

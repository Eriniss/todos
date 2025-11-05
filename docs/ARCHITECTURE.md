# Testbox Architecture Documentation

## System Overview

Testbox is designed as a learning platform to understand the transition from on-premise infrastructure to cloud-native AWS services. The architecture follows clean architecture principles with clear separation of concerns.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         Frontend                            │
│                  (HTML + CSS + JS)                          │
│                                                             │
│                   Nginx (Port 8080)                         │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      │ HTTP/HTTPS
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                    Backend (Go/Fiber v3)                    │
│                                                             │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │   Handler   │──│   Service    │──│   Repository     │  │
│  │    (API)    │  │  (Business)  │  │  (Data Access)   │  │
│  └─────────────┘  └──────┬───────┘  └────────┬─────────┘  │
│                           │                    │            │
│                           │                    │            │
│                    ┌──────▼────────┐     ┌────▼──────┐    │
│                    │  Cache Layer  │     │ Database  │    │
│                    │    (Redis)    │     │(PostgreSQL)│    │
│                    └───────────────┘     └───────────┘    │
│                           │                                 │
│                    ┌──────▼────────┐                       │
│                    │   Messaging   │                       │
│                    │  (RabbitMQ)   │                       │
│                    └───────────────┘                       │
└─────────────────────────────────────────────────────────────┘
                           │
                           │ (Future)
                           ▼
                    ┌──────────────┐
                    │   MongoDB    │
                    │  (Replicaset │
                    │  + Sharding) │
                    └──────────────┘
```

## Backend Architecture

### Layered Design

#### 1. API Layer (`internal/api`)
- **Responsibility**: HTTP request/response handling
- **Components**:
  - `handler.go`: HTTP handlers for CRUD operations
  - `routes.go`: Route definitions and middleware setup
- **Key Features**:
  - Input validation
  - Request binding
  - Error handling
  - HTTP status code management

#### 2. Service Layer (`internal/service`)
- **Responsibility**: Business logic and orchestration
- **Components**:
  - `todo_service.go`: Todo business logic
- **Key Features**:
  - Orchestrates repository, cache, and messaging
  - Implements caching strategies
  - Publishes events to message queue
  - Transaction management (future)

#### 3. Repository Layer (`internal/repository`)
- **Responsibility**: Data persistence abstraction
- **Components**:
  - `todo_repository.go`: Database operations
- **Key Features**:
  - Database abstraction
  - CRUD operations
  - Query building
  - Easy to swap implementations (PostgreSQL → DynamoDB)

#### 4. Cache Layer (`internal/cache`)
- **Responsibility**: Caching operations
- **Components**:
  - `redis.go`: Redis client and operations
- **Key Features**:
  - Lazy Loading pattern
  - Write-Through pattern
  - TTL management
  - Cache invalidation

#### 5. Messaging Layer (`internal/messaging`)
- **Responsibility**: Asynchronous event handling
- **Components**:
  - `rabbitmq.go`: Message queue operations
- **Key Features**:
  - Event publishing
  - Queue declaration
  - Message persistence
  - Prepared for SQS migration

## Caching Strategy Details

### Lazy Loading (Cache-Aside)

**Flow for Read Operations:**
```
1. Client requests todo by ID
2. Service checks Redis cache
3a. If CACHE HIT: Return cached data
3b. If CACHE MISS:
    → Fetch from PostgreSQL
    → Store in Redis with TTL
    → Return data to client
```

**Implementation:**
```go
func (s *todoService) GetTodo(id string) (*models.Todo, error) {
    // Try cache first
    cachedTodo, err := s.cache.GetTodo(id)
    if cachedTodo != nil {
        return cachedTodo, nil
    }

    // Cache miss - fetch from DB
    todo, err := s.repo.FindByID(id)

    // Populate cache for next time
    s.cache.SetTodo(todo)

    return todo, nil
}
```

**Benefits:**
- Only frequently accessed data is cached
- Reduces memory usage
- Cache grows organically based on access patterns

**Trade-offs:**
- Initial request is slower (cache miss)
- Potential for stale data if TTL is too long

### Write-Through

**Flow for Write Operations:**
```
1. Client creates/updates todo
2. Service writes to PostgreSQL
3. Service immediately writes to Redis
4. Service publishes event to RabbitMQ
5. Return success to client
```

**Implementation:**
```go
func (s *todoService) CreateTodo(title, content string) (*models.Todo, error) {
    // Write to database first
    if err := s.repo.Create(todo); err != nil {
        return nil, err
    }

    // Write-through to cache
    s.cache.SetTodo(todo)

    // Publish event
    s.rabbitmq.PublishEvent(messaging.TodoEvent{
        Action: "created",
        TodoID: todo.ID,
        Data:   todo,
    })

    return todo, nil
}
```

**Benefits:**
- Cache is always consistent with database
- No cache warming needed
- Read operations are fast

**Trade-offs:**
- Write latency is slightly higher
- Writes to cache even if data might not be read

## Database Design

### Todo Model

```sql
CREATE TABLE todos (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    completed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_todos_created_at ON todos(created_at DESC);
CREATE INDEX idx_todos_completed ON todos(completed);
```

### Migration to DynamoDB

**Planned Schema:**
```json
{
  "TableName": "testbox-todos",
  "KeySchema": [
    { "AttributeName": "id", "KeyType": "HASH" }
  ],
  "AttributeDefinitions": [
    { "AttributeName": "id", "AttributeType": "S" },
    { "AttributeName": "created_at", "AttributeType": "S" }
  ],
  "GlobalSecondaryIndexes": [
    {
      "IndexName": "CreatedAtIndex",
      "KeySchema": [
        { "AttributeName": "created_at", "KeyType": "HASH" }
      ]
    }
  ]
}
```

## Message Queue Design

### Event Structure

```go
type TodoEvent struct {
    Action string      `json:"action"` // "created", "updated", "deleted"
    TodoID string      `json:"todo_id"`
    Data   interface{} `json:"data,omitempty"`
}
```

### Use Cases

1. **Audit Logging**: Track all todo operations
2. **Analytics**: Count operations, track usage patterns
3. **Notifications**: Send notifications on todo completion
4. **Data Sync**: Sync to secondary databases (MongoDB)
5. **Cache Warming**: Pre-populate cache for related data

### Migration to SQS

**Planned Implementation:**
- Replace RabbitMQ client with AWS SDK
- Use FIFO queues for ordered processing
- Implement DLQ (Dead Letter Queue) for failed messages
- Add message attributes for filtering

## Scaling Considerations

### Horizontal Scaling

**Backend Pods/Instances:**
```
       ┌─────────────────┐
       │   Load Balancer │
       │      (ALB)      │
       └────────┬────────┘
                │
        ┌───────┴───────┐
        │               │
    ┌───▼───┐       ┌───▼───┐
    │Backend│       │Backend│
    │Pod 1  │       │Pod 2  │
    └───┬───┘       └───┬───┘
        │               │
        └───────┬───────┘
                │
        ┌───────▼────────┐
        │  Shared Cache  │
        │    (Redis)     │
        └────────────────┘
```

**Key Points:**
- Stateless backend design
- Shared Redis cache across all instances
- Session-less architecture
- Health checks for auto-scaling

### Database Scaling

**PostgreSQL:**
- Read replicas for read-heavy workloads
- Connection pooling
- Prepared for migration to DynamoDB

**MongoDB (Future):**
- Replica sets for high availability
- Sharding for horizontal scaling
- Read preference configuration

## Security Considerations

### Current Implementation

1. **Input Validation**: Basic validation in handlers
2. **XSS Prevention**: Frontend escapes HTML
3. **CORS**: Configured in Fiber middleware
4. **Error Handling**: Doesn't expose internal details

### Future Enhancements

1. **Authentication**: JWT-based auth
2. **Authorization**: RBAC for multi-tenant support
3. **Rate Limiting**: Protect against abuse
4. **Input Sanitization**: Comprehensive validation
5. **HTTPS**: TLS termination at load balancer
6. **Secrets Management**: AWS Secrets Manager
7. **Network Security**: VPC, Security Groups

## Observability

### Logging

**Current:**
- Basic console logging
- Operation tracking (cache hits/misses, DB operations)

**Planned:**
- Structured logging (JSON format)
- Log aggregation (CloudWatch, ELK Stack)
- Request ID tracking
- Correlation IDs for distributed tracing

### Monitoring

**Planned Metrics:**
- Request latency (p50, p95, p99)
- Cache hit rate
- Database query time
- Queue depth
- Error rate
- Active connections

**Tools:**
- Prometheus for metrics collection
- Grafana for visualization
- CloudWatch for AWS metrics
- X-Ray for distributed tracing

## Deployment Strategies

### Docker Compose (Current)

- Single host deployment
- Good for development and learning
- Easy to start and stop
- Not suitable for production

### Kubernetes (Planned)

**Resources:**
```yaml
- Deployments: Backend pods with multiple replicas
- Services: ClusterIP for internal, LoadBalancer for external
- ConfigMaps: Configuration management
- Secrets: Sensitive data
- PersistentVolumeClaims: Database storage
- HorizontalPodAutoscaler: Auto-scaling
- Ingress: External access
```

### AWS (Future)

**Option 1: ECS (Elastic Container Service)**
- Task definitions for backend
- ALB for load balancing
- Auto Scaling Groups
- RDS for PostgreSQL
- ElastiCache for Redis
- SQS for messaging

**Option 2: EKS (Elastic Kubernetes Service)**
- Kubernetes on AWS
- AWS CNI plugin
- ALB Ingress Controller
- EBS for persistent storage
- IAM for pod authentication

## Migration Path

### Phase 1: Containerization ✅
- Docker containers
- Docker Compose
- Multi-container orchestration

### Phase 2: Orchestration (In Progress)
- Kubernetes manifests
- Helm charts
- Local K8s testing (minikube/kind)

### Phase 3: Cloud Migration
- AWS ECS/EKS setup
- RDS PostgreSQL
- ElastiCache Redis
- SQS message queue
- DynamoDB exploration

### Phase 4: Optimization
- Auto-scaling policies
- Cost optimization
- Performance tuning
- Security hardening

## Learning Outcomes

By working with this architecture, you will learn:

1. **Backend Development**: Go, Fiber, GORM, clean architecture
2. **Caching Strategies**: Lazy loading, write-through patterns
3. **Message Queues**: Async processing, event-driven architecture
4. **Containerization**: Docker, multi-container apps
5. **Orchestration**: Kubernetes concepts and deployment
6. **Cloud Services**: AWS migration strategies
7. **Database Design**: SQL and NoSQL patterns
8. **Scalability**: Horizontal scaling, load balancing
9. **Observability**: Logging, monitoring, tracing
10. **DevOps**: CI/CD, infrastructure as code

---

This architecture is designed to be educational and incrementally complex, allowing you to learn each layer thoroughly before moving to the next level of complexity.

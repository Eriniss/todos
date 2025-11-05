# Testbox 아키텍처 문서

## 시스템 개요

Testbox는 온프레미스 인프라에서 클라우드 네이티브 AWS 서비스로의 전환을 이해하기 위한 학습 플랫폼으로 설계되었습니다. 아키텍처는 명확한 관심사 분리를 가진 클린 아키텍처 원칙을 따릅니다.

## 아키텍처 다이어그램

```
┌─────────────────────────────────────────────────────────────┐
│                         프론트엔드                           │
│                  (HTML + CSS + JS)                          │
│                                                             │
│                   Nginx (Port 8080)                         │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      │ HTTP/HTTPS
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                    백엔드 (Go/Fiber v2)                     │
│                                                             │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │   Handler   │──│   Service    │──│   Repository     │  │
│  │    (API)    │  │ (비즈니스)   │  │  (데이터 접근)   │  │
│  └─────────────┘  └──────┬───────┘  └────────┬─────────┘  │
│                           │                    │            │
│                           │                    │            │
│                    ┌──────▼────────┐     ┌────▼──────┐    │
│                    │  캐시 계층    │     │ 데이터베이스│    │
│                    │    (Redis)    │     │(PostgreSQL)│    │
│                    └───────────────┘     └───────────┘    │
│                           │                                 │
│                    ┌──────▼────────┐                       │
│                    │   메시징      │                       │
│                    │  (RabbitMQ)   │                       │
│                    └───────────────┘                       │
└─────────────────────────────────────────────────────────────┘
                           │
                           │ (향후)
                           ▼
                    ┌──────────────┐
                    │   MongoDB    │
                    │  (Replicaset │
                    │  + Sharding) │
                    └──────────────┘
```

## 백엔드 아키텍처

### 계층화된 설계

#### 1. API 계층 (`internal/api`)
- **책임**: HTTP 요청/응답 처리
- **컴포넌트**:
  - `handler.go`: CRUD 작업을 위한 HTTP 핸들러
  - `routes.go`: 라우트 정의 및 미들웨어 설정
- **주요 기능**:
  - 입력 검증
  - 요청 바인딩
  - 에러 처리
  - HTTP 상태 코드 관리

#### 2. 서비스 계층 (`internal/service`)
- **책임**: 비즈니스 로직 및 오케스트레이션
- **컴포넌트**:
  - `todo_service.go`: Todo 비즈니스 로직
- **주요 기능**:
  - Repository, Cache, Messaging 오케스트레이션
  - 캐싱 전략 구현
  - 메시지 큐에 이벤트 발행
  - 트랜잭션 관리 (향후)

#### 3. Repository 계층 (`internal/repository`)
- **책임**: 데이터 영속성 추상화
- **컴포넌트**:
  - `todo_repository.go`: 데이터베이스 작업
- **주요 기능**:
  - 데이터베이스 추상화
  - CRUD 작업
  - 쿼리 빌딩
  - 쉬운 구현 교체 (PostgreSQL → DynamoDB)

#### 4. 캐시 계층 (`internal/cache`)
- **책임**: 캐싱 작업
- **컴포넌트**:
  - `redis.go`: Redis 클라이언트 및 작업
- **주요 기능**:
  - Lazy Loading 패턴
  - Write-Through 패턴
  - TTL 관리
  - 캐시 무효화

#### 5. 메시징 계층 (`internal/messaging`)
- **책임**: 비동기 이벤트 처리
- **컴포넌트**:
  - `rabbitmq.go`: 메시지 큐 작업
- **주요 기능**:
  - 이벤트 발행
  - 큐 선언
  - 메시지 영속성
  - SQS 마이그레이션 준비 완료

## 캐싱 전략 상세

### Lazy Loading (Cache-Aside)

**읽기 작업 흐름:**
```
1. 클라이언트가 ID로 todo 요청
2. 서비스가 Redis 캐시 확인
3a. 캐시 HIT: 캐시된 데이터 반환
3b. 캐시 MISS:
    → PostgreSQL에서 조회
    → TTL과 함께 Redis에 저장
    → 클라이언트에 데이터 반환
```

**구현:**
```go
func (s *todoService) GetTodo(id string) (*models.Todo, error) {
    // 먼저 캐시 확인
    cachedTodo, err := s.cache.GetTodo(id)
    if cachedTodo != nil {
        return cachedTodo, nil
    }

    // 캐시 미스 - DB에서 조회
    todo, err := s.repo.FindByID(id)

    // 다음 조회를 위해 캐시에 저장
    s.cache.SetTodo(todo)

    return todo, nil
}
```

**장점:**
- 자주 접근하는 데이터만 캐시됨
- 메모리 사용량 감소
- 접근 패턴에 따라 캐시가 유기적으로 성장

**단점:**
- 초기 요청은 느림 (캐시 미스)
- TTL이 너무 길면 오래된 데이터 가능성

### Write-Through

**쓰기 작업 흐름:**
```
1. 클라이언트가 todo 생성/수정
2. 서비스가 PostgreSQL에 저장
3. 서비스가 즉시 Redis에 저장
4. 서비스가 RabbitMQ에 이벤트 발행
5. 클라이언트에 성공 응답 반환
```

**구현:**
```go
func (s *todoService) CreateTodo(title, content string) (*models.Todo, error) {
    // 먼저 데이터베이스에 저장
    if err := s.repo.Create(todo); err != nil {
        return nil, err
    }

    // Write-through로 캐시에 저장
    s.cache.SetTodo(todo)

    // 이벤트 발행
    s.rabbitmq.PublishEvent(messaging.TodoEvent{
        Action: "created",
        TodoID: todo.ID,
        Data:   todo,
    })

    return todo, nil
}
```

**장점:**
- 캐시가 항상 데이터베이스와 일관성 유지
- 캐시 워밍 불필요
- 읽기 작업이 빠름

**단점:**
- 쓰기 지연 시간이 약간 높음
- 읽지 않을 데이터도 캐시에 저장

## 데이터베이스 설계

### Todo 모델

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

### DynamoDB로의 마이그레이션

**계획된 스키마:**
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

## 메시지 큐 설계

### 이벤트 구조

```go
type TodoEvent struct {
    Action string      `json:"action"` // "created", "updated", "deleted"
    TodoID string      `json:"todo_id"`
    Data   interface{} `json:"data,omitempty"`
}
```

### 사용 사례

1. **감사 로깅**: 모든 todo 작업 추적
2. **분석**: 작업 횟수 집계, 사용 패턴 추적
3. **알림**: Todo 완료 시 알림 전송
4. **데이터 동기화**: 보조 데이터베이스(MongoDB)와 동기화
5. **캐시 워밍**: 관련 데이터를 위한 캐시 사전 로드

### SQS로의 마이그레이션

**계획된 구현:**
- RabbitMQ 클라이언트를 AWS SDK로 교체
- 순서 보장을 위한 FIFO 큐 사용
- 실패한 메시지를 위한 DLQ (Dead Letter Queue) 구현
- 필터링을 위한 메시지 속성 추가

## 확장 고려사항

### 수평 확장

**백엔드 Pod/인스턴스:**
```
       ┌─────────────────┐
       │  로드 밸런서     │
       │      (ALB)      │
       └────────┬────────┘
                │
        ┌───────┴───────┐
        │               │
    ┌───▼───┐       ┌───▼───┐
    │백엔드 │       │백엔드 │
    │Pod 1  │       │Pod 2  │
    └───┬───┘       └───┬───┘
        │               │
        └───────┬───────┘
                │
        ┌───────▼────────┐
        │  공유 캐시      │
        │    (Redis)     │
        └────────────────┘
```

**주요 포인트:**
- 무상태 백엔드 설계
- 모든 인스턴스에서 공유하는 Redis 캐시
- 세션리스 아키텍처
- 오토스케일링을 위한 헬스 체크

### 데이터베이스 확장

**PostgreSQL:**
- 읽기 중심 워크로드를 위한 Read Replica
- 커넥션 풀링
- DynamoDB 마이그레이션 준비 완료

**MongoDB (향후):**
- 고가용성을 위한 Replica Set
- 수평 확장을 위한 Sharding
- Read Preference 설정

## 보안 고려사항

### 현재 구현

1. **입력 검증**: 핸들러에서 기본 검증
2. **XSS 방지**: 프론트엔드에서 HTML 이스케이프
3. **CORS**: Fiber 미들웨어에서 설정
4. **에러 처리**: 내부 상세 정보 노출 방지

### 향후 개선 사항

1. **인증**: JWT 기반 인증
2. **권한 부여**: 멀티 테넌트 지원을 위한 RBAC
3. **속도 제한**: 남용 방지
4. **입력 검증**: 포괄적인 검증
5. **HTTPS**: 로드 밸런서에서 TLS 종료
6. **비밀 관리**: AWS Secrets Manager
7. **네트워크 보안**: VPC, Security Groups

## 관찰성 (Observability)

### 로깅

**현재:**
- 기본 콘솔 로깅
- 작업 추적 (캐시 히트/미스, DB 작업)

**계획:**
- 구조화된 로깅 (JSON 형식)
- 로그 집계 (CloudWatch, ELK Stack)
- Request ID 추적
- 분산 추적을 위한 Correlation ID

### 모니터링

**계획된 메트릭:**
- 요청 지연 시간 (p50, p95, p99)
- 캐시 히트율
- 데이터베이스 쿼리 시간
- 큐 깊이
- 에러율
- 활성 연결 수

**도구:**
- 메트릭 수집을 위한 Prometheus
- 시각화를 위한 Grafana
- AWS 메트릭을 위한 CloudWatch
- 분산 추적을 위한 X-Ray

## 배포 전략

### Docker Compose (현재)

- 단일 호스트 배포
- 개발 및 학습에 적합
- 시작 및 중지가 쉬움
- 프로덕션에는 적합하지 않음

### Kubernetes (계획)

**리소스:**
```yaml
- Deployments: 여러 replica를 가진 백엔드 Pod
- Services: 내부용 ClusterIP, 외부용 LoadBalancer
- ConfigMaps: 설정 관리
- Secrets: 민감한 데이터
- PersistentVolumeClaims: 데이터베이스 스토리지
- HorizontalPodAutoscaler: 오토스케일링
- Ingress: 외부 접근
```

### AWS (향후)

**옵션 1: ECS (Elastic Container Service)**
- 백엔드를 위한 Task Definition
- 로드 밸런싱을 위한 ALB
- Auto Scaling Groups
- PostgreSQL을 위한 RDS
- Redis를 위한 ElastiCache
- 메시징을 위한 SQS

**옵션 2: EKS (Elastic Kubernetes Service)**
- AWS 기반 Kubernetes
- AWS CNI 플러그인
- ALB Ingress Controller
- 영구 스토리지를 위한 EBS
- Pod 인증을 위한 IAM

## 마이그레이션 경로

### Phase 1: 컨테이너화 ✅
- Docker 컨테이너
- Docker Compose
- 멀티 컨테이너 오케스트레이션

### Phase 2: 오케스트레이션 (진행 중)
- Kubernetes 매니페스트
- Helm 차트
- 로컬 K8s 테스트 (minikube/kind)

### Phase 3: 클라우드 마이그레이션
- AWS ECS/EKS 설정
- RDS PostgreSQL
- ElastiCache Redis
- SQS 메시지 큐
- DynamoDB 탐색

### Phase 4: 최적화
- 오토스케일링 정책
- 비용 최적화
- 성능 튜닝
- 보안 강화

## 학습 성과

이 아키텍처를 작업하면서 다음을 배울 수 있습니다:

1. **백엔드 개발**: Go, Fiber, GORM, 클린 아키텍처
2. **캐싱 전략**: Lazy loading, Write-through 패턴
3. **메시지 큐**: 비동기 처리, 이벤트 기반 아키텍처
4. **컨테이너화**: Docker, 멀티 컨테이너 앱
5. **오케스트레이션**: Kubernetes 개념 및 배포
6. **클라우드 서비스**: AWS 마이그레이션 전략
7. **데이터베이스 설계**: SQL 및 NoSQL 패턴
8. **확장성**: 수평 확장, 로드 밸런싱
9. **관찰성**: 로깅, 모니터링, 추적
10. **DevOps**: CI/CD, Infrastructure as Code

---

이 아키텍처는 교육용으로 설계되었으며 점진적으로 복잡해지도록 구성되어 있어, 다음 복잡도 레벨로 넘어가기 전에 각 계층을 철저히 학습할 수 있습니다.

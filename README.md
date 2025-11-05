# Testbox - 온프레미스 & AWS 마이그레이션 테스트 플랫폼

온프레미스 서비스와 AWS 클라우드 마이그레이션 전략을 테스트하기 위한 종합 학습 프로젝트입니다. 백엔드 아키텍처, 캐싱 전략, 메시지 큐에 중점을 둔 간단한 Todo List 애플리케이션을 구현했습니다.

## 🎯 프로젝트 목적

- **교육용**: 다양한 온프레미스 및 AWS 서비스를 학습하고 실험
- **MVP**: 제목과 내용이 있는 간단한 Todo CRUD 작업
- **마이그레이션 준비**: 온프레미스에서 AWS 서비스로 쉽게 마이그레이션할 수 있도록 설계
- **확장 가능**: 로드 밸런싱, Kubernetes, MongoDB 샤딩을 지원하는 아키텍처

## 🏗️ 아키텍처

### 현재 스택 (온프레미스)

**백엔드:**
- Go 1.22 with Fiber v2 (HTTP 프레임워크)
- GORM (ORM)
- PostgreSQL (주 데이터베이스)
- Redis (Lazy Loading + Write-Through 캐싱 전략)
- RabbitMQ (비동기 작업을 위한 메시지 큐)
- MongoDB (향후 replicaset + sharding 테스트 준비)

**프론트엔드:**
- 바닐라 HTML, CSS, JavaScript
- Nginx (웹 서버 및 리버스 프록시)

**인프라:**
- Docker & Docker Compose
- Kubernetes 마이그레이션 준비 완료

### 향후 마이그레이션 경로 (AWS)

- **PostgreSQL** → **DynamoDB**
- **RabbitMQ** → **SQS**
- **Fiber 백엔드** → **Elastic Beanstalk** 또는 **ECS/EKS**
- **로드 밸런싱** → **ALB + Auto Scaling Groups**
- **컴퓨팅** → **EC2** 또는 **ECS/EKS**

## 📁 프로젝트 구조

```
testbox/
├── backend/                 # Go/Fiber 백엔드
│   ├── cmd/
│   │   └── main.go         # 애플리케이션 진입점
│   ├── internal/
│   │   ├── api/            # HTTP 핸들러 및 라우트
│   │   ├── cache/          # Redis 캐싱 구현
│   │   ├── config/         # 설정 관리
│   │   ├── database/       # 데이터베이스 연결
│   │   ├── messaging/      # RabbitMQ 통합
│   │   ├── models/         # 데이터 모델
│   │   ├── repository/     # 데이터 접근 계층
│   │   └── service/        # 비즈니스 로직 계층
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── frontend/               # 바닐라 JS 프론트엔드
│   ├── css/
│   │   └── style.css
│   ├── js/
│   │   └── app.js
│   └── index.html
├── docs/                   # 문서
├── docker-compose.yml      # 멀티 컨테이너 오케스트레이션
├── nginx.conf             # Nginx 설정
├── Makefile               # 개발 편의 명령어
└── README.md
```

## 🚀 빠른 시작

### 사전 요구사항

- Docker & Docker Compose
- (선택) Go 1.22+ (로컬 개발용)
- (선택) Make (편의 명령어용)

### Docker Compose로 실행 (권장)

1. **프로젝트 디렉토리로 이동:**
   ```bash
   cd testbox
   ```

2. **모든 서비스 시작:**
   ```bash
   make up
   # 또는
   docker-compose up -d
   ```

3. **애플리케이션 접속:**
   - 프론트엔드: http://localhost:8080
   - 백엔드 API: http://localhost:3000/api/todos
   - 헬스 체크: http://localhost:3000/health
   - RabbitMQ 관리: http://localhost:15672 (guest/guest)

4. **로그 확인:**
   ```bash
   make logs
   # 또는
   docker-compose logs -f
   ```

5. **서비스 중지:**
   ```bash
   make down
   # 또는
   docker-compose down
   ```

### 로컬 개발 모드

1. **인프라 서비스만 시작:**
   ```bash
   docker-compose up -d postgres redis rabbitmq mongodb
   ```

2. **백엔드 실행:**
   ```bash
   cd backend
   go mod download
   go run cmd/main.go
   ```

3. **프론트엔드 서빙:**
   ```bash
   cd frontend
   python3 -m http.server 8080
   # 또는
   npx serve -p 8080
   ```

## 🔧 설정

`.env` 파일에서 환경 변수를 설정할 수 있습니다 (`.env.example` 참조):

```env
SERVER_PORT=3000
DB_HOST=localhost
DB_PORT=5432
REDIS_HOST=localhost
REDIS_PORT=6379
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
MONGO_URL=mongodb://localhost:27017
```

## 📚 주요 학습 포인트

### 1. 캐싱 전략

**Lazy Loading (Cache-Aside):**
- `GetTodo()` 작업에서 사용
- 캐시 먼저 확인 → 미스 시 DB 조회 → 캐시 저장
- 장점: 필요한 데이터만 캐싱

**Write-Through:**
- `CreateTodo()` 및 `UpdateTodo()` 작업에서 사용
- DB에 먼저 저장 → 즉시 캐시에 저장
- 장점: 캐시가 항상 최신 상태 유지

구현: [backend/internal/cache/redis.go](backend/internal/cache/redis.go)

### 2. 계층화된 아키텍처

```
Handler (API) → Service (비즈니스 로직) → Repository (데이터 접근)
                    ↓                           ↓
                  Cache                      Database
                    ↓
                Messaging
```

### 3. 메시지 큐 통합

- Todo 작업에 대한 비동기 이벤트 발행
- AWS SQS로 마이그레이션 준비 완료
- 구현: [backend/internal/messaging/rabbitmq.go](backend/internal/messaging/rabbitmq.go)

### 4. 데이터베이스 추상화

- 쉬운 데이터베이스 전환을 위한 Repository 패턴
- GORM을 사용한 ORM 작업
- DynamoDB 마이그레이션 준비 완료

## 🔍 API 엔드포인트

| 메서드 | 엔드포인트 | 설명 |
|--------|----------|-------------|
| GET | `/api/todos` | 모든 todo 조회 |
| GET | `/api/todos/:id` | 특정 todo 조회 |
| POST | `/api/todos` | 새 todo 생성 |
| PUT | `/api/todos/:id` | todo 수정 |
| DELETE | `/api/todos/:id` | todo 삭제 |
| GET | `/health` | 헬스 체크 |

### 예시 요청

**Todo 생성:**
```bash
curl -X POST http://localhost:3000/api/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Go 학습하기","content":"Fiber v2 프레임워크 공부"}'
```

**전체 Todo 조회:**
```bash
curl http://localhost:3000/api/todos
```

**Todo 수정:**
```bash
curl -X PUT http://localhost:3000/api/todos/{id} \
  -H "Content-Type: application/json" \
  -d '{"title":"Go 학습하기","content":"완료!","completed":true}'
```

## 🎓 학습 로드맵

### Phase 1: 온프레미스 (현재)
- ✅ Go/Fiber 백엔드 with GORM
- ✅ PostgreSQL 데이터베이스
- ✅ Redis 캐싱 (Lazy Loading + Write-Through)
- ✅ RabbitMQ 메시지 큐
- ✅ Docker 컨테이너화
- 🔄 Kubernetes 배포 (예정)
- 🔄 MongoDB replicaset + sharding (예정)

### Phase 2: AWS 마이그레이션
- 🔄 DynamoDB 통합
- 🔄 SQS 메시지 큐
- 🔄 Elastic Beanstalk 배포
- 🔄 ALB + Auto Scaling Groups
- 🔄 ECS/EKS 컨테이너 오케스트레이션

### Phase 3: 고급 주제
- 🔄 다중 리전 배포
- 🔄 Blue-green 배포
- 🔄 Canary 릴리스
- 🔄 성능 테스트 및 최적화
- 🔄 관찰성 (로깅, 모니터링, 추적)

## 🛠️ Makefile 명령어

```bash
make help           # 사용 가능한 모든 명령어 표시
make up             # 모든 서비스 시작
make down           # 모든 서비스 중지
make build          # 모든 서비스 빌드
make logs           # 로그 표시
make clean          # 모든 컨테이너 및 볼륨 제거
make restart        # 서비스 재시작
make health         # 모든 서비스의 상태 확인
make backend-dev    # 백엔드를 개발 모드로 실행
make frontend-dev   # 프론트엔드를 로컬로 서빙
```

## 📊 서비스 포트

| 서비스 | 포트 | 설명 |
|---------|------|-------------|
| 프론트엔드 | 8080 | Nginx 웹 서버 |
| 백엔드 | 3000 | Go/Fiber API |
| PostgreSQL | 5432 | 데이터베이스 |
| Redis | 6379 | 캐시 |
| RabbitMQ | 5672 | 메시지 큐 |
| RabbitMQ UI | 15672 | 관리 인터페이스 |
| MongoDB | 27017 | 문서 데이터베이스 |

## 🧪 테스트

```bash
# 백엔드 테스트
cd backend
go test -v ./...

# 헬스 체크
make health
```

## 📝 향후 개선 사항

- [ ] 포괄적인 단위 및 통합 테스트 추가
- [ ] PostgreSQL과 함께 MongoDB 어댑터 구현
- [ ] Kubernetes 매니페스트 생성 (deployments, services, ingress)
- [ ] CI/CD 파이프라인 추가
- [ ] DynamoDB 및 SQS를 위한 AWS SDK 구현
- [ ] 관찰성 스택 추가 (Prometheus, Grafana)
- [ ] 인증 및 권한 부여 구현
- [ ] 속도 제한 및 요청 스로틀링 추가
- [ ] AWS를 위한 Terraform/CloudFormation 템플릿 생성

## 🤝 기여

개인 학습 프로젝트이지만 제안 및 개선 사항은 환영합니다!

## 📄 라이선스

MIT License - 학습 목적으로 자유롭게 사용하세요.

---

**참고:** 이 프로젝트는 학습 및 실험용으로 설계되었습니다. 프로덕션 사용 전에 인증, 입력 검증, 속도 제한과 같은 보안 기능을 강화해야 합니다.

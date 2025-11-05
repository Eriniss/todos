# Getting Started with Testbox

## 빠른 시작 가이드

### 전제 조건

- Docker & Docker Compose 설치
- (선택) Go 1.22+ (로컬 개발용)
- (선택) Make (편의 명령어용)

### 1. Docker Compose로 전체 실행 (추천)

가장 간단한 방법입니다. 모든 서비스를 한 번에 실행합니다.

```bash
# 1. 프로젝트 디렉토리로 이동
cd testbox

# 2. 모든 서비스 시작
make up
# 또는
docker-compose up -d

# 3. 로그 확인
make logs
# 또는
docker-compose logs -f

# 4. 서비스 접속
# - 프론트엔드: http://localhost:8080
# - 백엔드 API: http://localhost:3000/api/todos
# - Health Check: http://localhost:3000/health
# - RabbitMQ 관리: http://localhost:15672 (guest/guest)
```

### 2. 로컬 개발 모드

백엔드 코드를 수정하면서 개발하려면 이 방법을 사용합니다.

```bash
# 1. 인프라 서비스만 시작
docker-compose up -d postgres redis rabbitmq mongodb

# 2. 백엔드 로컬 실행
cd backend
go mod download
go run cmd/main.go

# 3. 다른 터미널에서 프론트엔드 서빙
cd frontend
python3 -m http.server 8080
# 또는
npx serve -p 8080
```

### 3. 서비스 상태 확인

```bash
# 실행 중인 컨테이너 확인
docker-compose ps

# 헬스 체크
make health

# 개별 서비스 로그
docker-compose logs backend
docker-compose logs postgres
docker-compose logs redis
docker-compose logs rabbitmq
```

## 주요 명령어

### Make 명령어

```bash
make help           # 사용 가능한 명령어 목록
make up             # 모든 서비스 시작
make down           # 모든 서비스 중지
make build          # 모든 서비스 빌드
make logs           # 전체 로그 보기
make restart        # 서비스 재시작
make clean          # 모든 컨테이너, 볼륨, 이미지 삭제
make health         # 서비스 헬스 체크
make backend-dev    # 백엔드 로컬 실행
make frontend-dev   # 프론트엔드 로컬 실행
```

### Docker Compose 명령어

```bash
# 서비스 시작
docker-compose up -d

# 특정 서비스만 시작
docker-compose up -d backend redis

# 서비스 중지
docker-compose down

# 볼륨까지 삭제
docker-compose down -v

# 로그 보기
docker-compose logs -f [서비스명]

# 서비스 재시작
docker-compose restart [서비스명]

# 컨테이너 상태
docker-compose ps
```

## API 테스트

### cURL로 API 테스트

```bash
# Health Check
curl http://localhost:3000/health

# 전체 Todo 조회
curl http://localhost:3000/api/todos

# Todo 생성
curl -X POST http://localhost:3000/api/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Go","content":"Study Fiber framework"}'

# Todo 조회 (ID 필요)
curl http://localhost:3000/api/todos/{id}

# Todo 수정
curl -X PUT http://localhost:3000/api/todos/{id} \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Go","content":"Completed!","completed":true}'

# Todo 삭제
curl -X DELETE http://localhost:3000/api/todos/{id}
```

### 프론트엔드에서 테스트

1. 브라우저에서 http://localhost:8080 접속
2. "Add New Todo" 폼에서 제목과 내용 입력
3. "Add Todo" 버튼 클릭
4. Todo 목록에서 생성된 항목 확인
5. "Edit" 버튼으로 수정 가능
6. "Delete" 버튼으로 삭제 가능

## 문제 해결

### 포트 충돌

다른 애플리케이션이 동일한 포트를 사용 중인 경우:

```bash
# 사용 중인 포트 확인
lsof -i :3000
lsof -i :8080
lsof -i :5432

# 프로세스 종료
kill -9 <PID>

# 또는 docker-compose.yml에서 포트 변경
```

### 데이터베이스 연결 실패

```bash
# PostgreSQL 로그 확인
docker-compose logs postgres

# 데이터베이스 재시작
docker-compose restart postgres

# 데이터베이스 리셋 (주의: 모든 데이터 삭제)
make db-reset
```

### Redis 연결 실패

```bash
# Redis 로그 확인
docker-compose logs redis

# Redis 재시작
docker-compose restart redis

# Redis 연결 테스트
docker-compose exec redis redis-cli ping
```

### RabbitMQ 연결 실패

```bash
# RabbitMQ 로그 확인
docker-compose logs rabbitmq

# RabbitMQ 재시작
docker-compose restart rabbitmq

# RabbitMQ 관리 UI 접속
open http://localhost:15672
# 계정: guest / guest
```

### 백엔드 빌드 실패

```bash
# Go 의존성 재다운로드
cd backend
rm go.sum
go mod download
go mod tidy

# 빌드 테스트
go build -o testbox ./cmd/main.go
```

### 프론트엔드가 백엔드에 연결 안됨

1. 백엔드가 실행 중인지 확인: `curl http://localhost:3000/health`
2. CORS 설정 확인: [backend/cmd/main.go](../backend/cmd/main.go) 참조
3. 프론트엔드 API URL 확인: [frontend/js/app.js](../frontend/js/app.js) 의 `API_BASE_URL`

### 전체 리셋

모든 것을 처음부터 다시 시작하려면:

```bash
# 모든 컨테이너, 볼륨, 이미지 삭제
make clean
# 또는
docker-compose down -v --rmi all

# 다시 시작
make up
```

## 캐싱 동작 확인

백엔드 로그에서 캐싱 동작을 확인할 수 있습니다:

```bash
# 백엔드 로그 모니터링
docker-compose logs -f backend

# Todo 생성 시
✓ Created todo: {id}
✓ Cache SET: todo:{id}

# Todo 조회 시 (첫 번째 - Cache Miss)
Cache MISS: todo:{id} - fetching from DB
✓ Cache SET: todo:{id}

# Todo 조회 시 (두 번째 - Cache Hit)
✓ Cache HIT: todo:{id}

# Todo 수정 시
✓ Updated todo: {id}
✓ Cache SET: todo:{id}

# Todo 삭제 시
✓ Deleted todo: {id}
✓ Cache DELETE: todo:{id}
```

## 다음 단계

프로젝트가 정상적으로 실행되면:

1. **코드 탐색**: [ARCHITECTURE.md](./ARCHITECTURE.md)에서 아키텍처 이해
2. **캐싱 전략 학습**: [backend/internal/cache/redis.go](../backend/internal/cache/redis.go) 분석
3. **비즈니스 로직 학습**: [backend/internal/service/todo_service.go](../backend/internal/service/todo_service.go) 분석
4. **Kubernetes 배포**: K8s 매니페스트 작성 (향후)
5. **AWS 마이그레이션**: DynamoDB, SQS로 전환 (향후)

## 참고 자료

- [프로젝트 README](../README.md)
- [아키텍처 문서](./ARCHITECTURE.md)
- [Fiber v2 공식 문서](https://docs.gofiber.io/)
- [GORM 공식 문서](https://gorm.io/)
- [Redis 공식 문서](https://redis.io/docs/)
- [RabbitMQ 공식 문서](https://www.rabbitmq.com/documentation.html)

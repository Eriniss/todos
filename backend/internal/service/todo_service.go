package service

import (
	"fmt"
	"log"
	"testbox/internal/cache"
	"testbox/internal/messaging"
	"testbox/internal/models"
	"testbox/internal/repository"

	"gorm.io/gorm"
)

type TodoService interface {
	CreateTodo(title, content string) (*models.Todo, error)
	GetTodo(id string) (*models.Todo, error)
	GetAllTodos() ([]models.Todo, error)
	UpdateTodo(id, title, content string, completed bool) (*models.Todo, error)
	DeleteTodo(id string) error
}

type todoService struct {
	repo     repository.TodoRepository
	cache    *cache.RedisCache
	rabbitmq *messaging.RabbitMQ
}

func NewTodoService(repo repository.TodoRepository, cache *cache.RedisCache, rabbitmq *messaging.RabbitMQ) TodoService {
	return &todoService{
		repo:     repo,
		cache:    cache,
		rabbitmq: rabbitmq,
	}
}

// CreateTodo 는 Write-Through 캐싱 전략을 구현합니다
func (s *todoService) CreateTodo(title, content string) (*models.Todo, error) {
	todo := &models.Todo{
		Title:   title,
		Content: content,
	}

	// 1. 먼저 데이터베이스에 저장
	if err := s.repo.Create(todo); err != nil {
		return nil, fmt.Errorf("Todo 생성 실패: %w", err)
	}

	// 2. Write-Through: 즉시 캐시에 저장
	if err := s.cache.SetTodo(todo); err != nil {
		log.Printf("경고: 캐시 저장 실패: %v", err)
		// 캐시 실패는 치명적이지 않으므로 요청은 성공 처리
	}

	// 3. 비동기 처리를 위한 이벤트 발행
	if err := s.rabbitmq.PublishEvent(messaging.TodoEvent{
		Action: "created",
		TodoID: todo.ID,
		Data:   todo,
	}); err != nil {
		log.Printf("경고: 이벤트 발행 실패: %v", err)
	}

	log.Printf("✓ Todo 생성 완료: %s", todo.ID)
	return todo, nil
}

// GetTodo 는 Lazy Loading 캐싱 전략을 구현합니다
func (s *todoService) GetTodo(id string) (*models.Todo, error) {
	// 1. 먼저 캐시 확인 (Lazy Loading)
	cachedTodo, err := s.cache.GetTodo(id)
	if err != nil {
		log.Printf("경고: 캐시 오류: %v", err)
	}
	if cachedTodo != nil {
		return cachedTodo, nil
	}

	// 2. 캐시 미스 - 데이터베이스에서 조회
	log.Printf("캐시 미스: todo:%s - DB에서 조회 중", id)
	todo, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("Todo를 찾을 수 없습니다")
		}
		return nil, fmt.Errorf("Todo 조회 실패: %w", err)
	}

	// 3. Lazy Loading: 다음 조회를 위해 캐시에 저장
	if err := s.cache.SetTodo(todo); err != nil {
		log.Printf("경고: 캐시 업데이트 실패: %v", err)
	}

	return todo, nil
}

// GetAllTodos 는 모든 Todo 목록을 조회합니다
func (s *todoService) GetAllTodos() ([]models.Todo, error) {
	// 목록 조회는 데이터베이스에서 직접 조회합니다
	// 리스트 캐싱은 복잡하고 이 사례에서는 효율적이지 않습니다
	todos, err := s.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("Todo 목록 조회 실패: %w", err)
	}
	return todos, nil
}

// UpdateTodo 는 Write-Through 캐싱 전략을 구현합니다
func (s *todoService) UpdateTodo(id, title, content string, completed bool) (*models.Todo, error) {
	// 1. 기존 Todo 조회
	todo, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("Todo를 찾을 수 없습니다")
		}
		return nil, fmt.Errorf("Todo 조회 실패: %w", err)
	}

	// 2. 필드 업데이트
	todo.Title = title
	todo.Content = content
	todo.Completed = completed

	// 3. 데이터베이스에 저장
	if err := s.repo.Update(todo); err != nil {
		return nil, fmt.Errorf("Todo 업데이트 실패: %w", err)
	}

	// 4. Write-Through: 즉시 캐시 업데이트
	if err := s.cache.SetTodo(todo); err != nil {
		log.Printf("경고: 캐시 업데이트 실패: %v", err)
	}

	// 5. 이벤트 발행
	if err := s.rabbitmq.PublishEvent(messaging.TodoEvent{
		Action: "updated",
		TodoID: todo.ID,
		Data:   todo,
	}); err != nil {
		log.Printf("경고: 이벤트 발행 실패: %v", err)
	}

	log.Printf("✓ Todo 업데이트 완료: %s", todo.ID)
	return todo, nil
}

// DeleteTodo 는 Todo를 삭제하고 캐시를 무효화합니다
func (s *todoService) DeleteTodo(id string) error {
	// 1. 데이터베이스에서 삭제
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("Todo 삭제 실패: %w", err)
	}

	// 2. 캐시 무효화
	if err := s.cache.DeleteTodo(id); err != nil {
		log.Printf("경고: 캐시 삭제 실패: %v", err)
	}

	// 3. 이벤트 발행
	if err := s.rabbitmq.PublishEvent(messaging.TodoEvent{
		Action: "deleted",
		TodoID: id,
	}); err != nil {
		log.Printf("경고: 이벤트 발행 실패: %v", err)
	}

	log.Printf("✓ Todo 삭제 완료: %s", id)
	return nil
}

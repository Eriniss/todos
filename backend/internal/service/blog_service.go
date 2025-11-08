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

type BlogService interface {
	CreateBlogPost(title, content, tags string) (*models.BlogPost, error)
	GetBlogPost(id string) (*models.BlogPost, error)
	GetAllBlogPosts() ([]models.BlogPost, error)
	UpdateBlogPost(id, title, content, tags string) (*models.BlogPost, error)
	DeleteBlogPost(id string) error
}

type blogService struct {
	repo     repository.BlogRepository
	cache    *cache.RedisCache
	rabbitmq *messaging.RabbitMQ
}

func NewBlogService(repo repository.BlogRepository, cache *cache.RedisCache, rabbitmq *messaging.RabbitMQ) BlogService {
	return &blogService{
		repo:     repo,
		cache:    cache,
		rabbitmq: rabbitmq,
	}
}

// CreateBlogPost creates a new blog post
func (s *blogService) CreateBlogPost(title, content, tags string) (*models.BlogPost, error) {
	blog := &models.BlogPost{
		Title:   title,
		Content: content,
		Tags:    tags,
	}

	// Save to database
	if err := s.repo.Create(blog); err != nil {
		return nil, fmt.Errorf("블로그 포스트 생성 실패: %w", err)
	}

	// Publish event for async processing
	if err := s.rabbitmq.PublishEvent(messaging.TodoEvent{
		Action: "blog_created",
		TodoID: blog.ID,
		Data:   blog,
	}); err != nil {
		log.Printf("경고: 이벤트 발행 실패: %v", err)
	}

	log.Printf("✓ 블로그 포스트 생성 완료: %s", blog.ID)
	return blog, nil
}

// GetBlogPost retrieves a blog post by ID
func (s *blogService) GetBlogPost(id string) (*models.BlogPost, error) {
	blog, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("블로그 포스트를 찾을 수 없습니다")
		}
		return nil, fmt.Errorf("블로그 포스트 조회 실패: %w", err)
	}

	return blog, nil
}

// GetAllBlogPosts retrieves all blog posts
func (s *blogService) GetAllBlogPosts() ([]models.BlogPost, error) {
	blogs, err := s.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("블로그 포스트 목록 조회 실패: %w", err)
	}
	return blogs, nil
}

// UpdateBlogPost updates an existing blog post
func (s *blogService) UpdateBlogPost(id, title, content, tags string) (*models.BlogPost, error) {
	// Find existing blog post
	blog, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("블로그 포스트를 찾을 수 없습니다")
		}
		return nil, fmt.Errorf("블로그 포스트 조회 실패: %w", err)
	}

	// Update fields
	blog.Title = title
	blog.Content = content
	blog.Tags = tags

	// Save to database
	if err := s.repo.Update(blog); err != nil {
		return nil, fmt.Errorf("블로그 포스트 업데이트 실패: %w", err)
	}

	// Publish event
	if err := s.rabbitmq.PublishEvent(messaging.TodoEvent{
		Action: "blog_updated",
		TodoID: blog.ID,
		Data:   blog,
	}); err != nil {
		log.Printf("경고: 이벤트 발행 실패: %v", err)
	}

	log.Printf("✓ 블로그 포스트 업데이트 완료: %s", blog.ID)
	return blog, nil
}

// DeleteBlogPost deletes a blog post
func (s *blogService) DeleteBlogPost(id string) error {
	// Delete from database
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("블로그 포스트 삭제 실패: %w", err)
	}

	// Publish event
	if err := s.rabbitmq.PublishEvent(messaging.TodoEvent{
		Action: "blog_deleted",
		TodoID: id,
	}); err != nil {
		log.Printf("경고: 이벤트 발행 실패: %v", err)
	}

	log.Printf("✓ 블로그 포스트 삭제 완료: %s", id)
	return nil
}

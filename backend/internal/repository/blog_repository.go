package repository

import (
	"testbox/internal/models"

	"gorm.io/gorm"
)

type BlogRepository interface {
	Create(blog *models.BlogPost) error
	FindByID(id string) (*models.BlogPost, error)
	FindAll() ([]models.BlogPost, error)
	Update(blog *models.BlogPost) error
	Delete(id string) error
}

type blogRepository struct {
	db *gorm.DB
}

func NewBlogRepository(db *gorm.DB) BlogRepository {
	return &blogRepository{db: db}
}

func (r *blogRepository) Create(blog *models.BlogPost) error {
	return r.db.Create(blog).Error
}

func (r *blogRepository) FindByID(id string) (*models.BlogPost, error) {
	var blog models.BlogPost
	if err := r.db.First(&blog, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &blog, nil
}

func (r *blogRepository) FindAll() ([]models.BlogPost, error) {
	var blogs []models.BlogPost
	if err := r.db.Order("created_at DESC").Find(&blogs).Error; err != nil {
		return nil, err
	}
	return blogs, nil
}

func (r *blogRepository) Update(blog *models.BlogPost) error {
	return r.db.Save(blog).Error
}

func (r *blogRepository) Delete(id string) error {
	return r.db.Delete(&models.BlogPost{}, "id = ?", id).Error
}

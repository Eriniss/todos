package repository

import (
	"testbox/internal/models"

	"gorm.io/gorm"
)

type TodoRepository interface {
	Create(todo *models.Todo) error
	FindByID(id string) (*models.Todo, error)
	FindAll() ([]models.Todo, error)
	Update(todo *models.Todo) error
	Delete(id string) error
}

type todoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) Create(todo *models.Todo) error {
	return r.db.Create(todo).Error
}

func (r *todoRepository) FindByID(id string) (*models.Todo, error) {
	var todo models.Todo
	if err := r.db.First(&todo, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &todo, nil
}

func (r *todoRepository) FindAll() ([]models.Todo, error) {
	var todos []models.Todo
	if err := r.db.Order("created_at DESC").Find(&todos).Error; err != nil {
		return nil, err
	}
	return todos, nil
}

func (r *todoRepository) Update(todo *models.Todo) error {
	return r.db.Save(todo).Error
}

func (r *todoRepository) Delete(id string) error {
	return r.db.Delete(&models.Todo{}, "id = ?", id).Error
}

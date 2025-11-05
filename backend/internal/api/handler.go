package api

import (
	"testbox/internal/service"

	"github.com/gofiber/fiber/v2"
)

type TodoHandler struct {
	service service.TodoService
}

func NewTodoHandler(service service.TodoService) *TodoHandler {
	return &TodoHandler{service: service}
}

type CreateTodoRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdateTodoRequest struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	Completed bool   `json:"completed"`
}

// CreateTodo 는 새로운 Todo를 생성합니다
// @Summary 새로운 Todo 생성
// @Description 제목과 내용을 받아 새로운 Todo를 생성합니다
// @Tags todos
// @Accept json
// @Produce json
// @Success 201 {object} models.Todo
// @Router /api/todos [post]
func (h *TodoHandler) CreateTodo(c *fiber.Ctx) error {
	var req CreateTodoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title is required",
		})
	}

	todo, err := h.service.CreateTodo(req.Title, req.Content)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(todo)
}

// GetTodo 는 ID로 특정 Todo를 조회합니다
// @Summary ID로 Todo 조회
// @Description 주어진 ID에 해당하는 Todo를 조회합니다 (캐시 우선 조회)
// @Tags todos
// @Produce json
// @Param id path string true "Todo ID"
// @Success 200 {object} models.Todo
// @Router /api/todos/{id} [get]
func (h *TodoHandler) GetTodo(c *fiber.Ctx) error {
	id := c.Params("id")

	todo, err := h.service.GetTodo(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(todo)
}

// GetAllTodos 는 모든 Todo 목록을 조회합니다
// @Summary 전체 Todo 목록 조회
// @Description 등록된 모든 Todo를 생성일 역순으로 조회합니다
// @Tags todos
// @Produce json
// @Success 200 {array} models.Todo
// @Router /api/todos [get]
func (h *TodoHandler) GetAllTodos(c *fiber.Ctx) error {
	todos, err := h.service.GetAllTodos()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(todos)
}

// UpdateTodo 는 기존 Todo를 수정합니다
// @Summary Todo 수정
// @Description 제목, 내용, 완료 상태를 수정하고 캐시를 업데이트합니다 (Write-Through)
// @Tags todos
// @Accept json
// @Produce json
// @Param id path string true "Todo ID"
// @Success 200 {object} models.Todo
// @Router /api/todos/{id} [put]
func (h *TodoHandler) UpdateTodo(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateTodoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title is required",
		})
	}

	todo, err := h.service.UpdateTodo(id, req.Title, req.Content, req.Completed)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(todo)
}

// DeleteTodo 는 Todo를 삭제합니다
// @Summary Todo 삭제
// @Description Todo를 삭제하고 캐시에서도 제거합니다
// @Tags todos
// @Param id path string true "Todo ID"
// @Success 204
// @Router /api/todos/{id} [delete]
func (h *TodoHandler) DeleteTodo(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.DeleteTodo(id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

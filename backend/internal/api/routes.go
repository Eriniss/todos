package api

import (
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes 는 애플리케이션의 모든 라우트를 설정합니다
func SetupRoutes(app *fiber.App, handler *TodoHandler) {
	// API 라우트 그룹
	api := app.Group("/api")

	// Todo 관련 라우트
	todos := api.Group("/todos")
	todos.Post("/", handler.CreateTodo)      // Todo 생성
	todos.Get("/", handler.GetAllTodos)      // 전체 Todo 조회
	todos.Get("/:id", handler.GetTodo)       // 특정 Todo 조회
	todos.Put("/:id", handler.UpdateTodo)    // Todo 수정
	todos.Delete("/:id", handler.DeleteTodo) // Todo 삭제

	// 헬스 체크 엔드포인트
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "testbox-backend",
		})
	})
}

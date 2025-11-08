package api

import (
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes 는 애플리케이션의 모든 라우트를 설정합니다
func SetupRoutes(app *fiber.App, todoHandler *TodoHandler, blogHandler *BlogHandler) {
	// API 라우트 그룹
	api := app.Group("/api")

	// Todo 관련 라우트
	todos := api.Group("/todos")
	todos.Post("/", todoHandler.CreateTodo)      // Todo 생성
	todos.Get("/", todoHandler.GetAllTodos)      // 전체 Todo 조회
	todos.Get("/:id", todoHandler.GetTodo)       // 특정 Todo 조회
	todos.Put("/:id", todoHandler.UpdateTodo)    // Todo 수정
	todos.Delete("/:id", todoHandler.DeleteTodo) // Todo 삭제

	// Blog 관련 라우트
	blogs := api.Group("/blogs")
	blogs.Post("/", blogHandler.CreateBlog)      // Blog 생성
	blogs.Get("/", blogHandler.GetAllBlogs)      // 전체 Blog 조회
	blogs.Get("/:id", blogHandler.GetBlog)       // 특정 Blog 조회
	blogs.Put("/:id", blogHandler.UpdateBlog)    // Blog 수정
	blogs.Delete("/:id", blogHandler.DeleteBlog) // Blog 삭제

	// 헬스 체크 엔드포인트
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "testbox-backend",
		})
	})
}

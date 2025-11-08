package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testbox/internal/api"
	"testbox/internal/cache"
	"testbox/internal/config"
	"testbox/internal/database"
	"testbox/internal/messaging"
	"testbox/internal/repository"
	"testbox/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// 설정 로드
	cfg := config.LoadConfig()

	// PostgreSQL 초기화
	postgresDB, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("PostgreSQL 초기화 실패: %v", err)
	}
	defer postgresDB.Close()

	// Redis 캐시 초기화
	redisCache, err := cache.NewRedisCache(cfg)
	if err != nil {
		log.Fatalf("Redis 초기화 실패: %v", err)
	}
	defer redisCache.Close()

	// RabbitMQ 초기화
	rabbitMQ, err := messaging.NewRabbitMQ(cfg)
	if err != nil {
		log.Fatalf("RabbitMQ 초기화 실패: %v", err)
	}
	defer rabbitMQ.Close()

	// MongoDB 초기화 (향후 사용 예정)
	mongoDB, err := database.NewMongoDB(cfg)
	if err != nil {
		log.Printf("경고: MongoDB 초기화 실패: %v", err)
		// MongoDB는 현재 선택사항입니다
	} else {
		defer mongoDB.Close()
	}

	// 각 레이어 초기화
	todoRepo := repository.NewTodoRepository(postgresDB.DB)
	todoService := service.NewTodoService(todoRepo, redisCache, rabbitMQ)
	todoHandler := api.NewTodoHandler(todoService)

	blogRepo := repository.NewBlogRepository(postgresDB.DB)
	blogService := service.NewBlogService(blogRepo, redisCache, rabbitMQ)
	blogHandler := api.NewBlogHandler(blogService)

	// Fiber 앱 생성
	app := fiber.New(fiber.Config{
		AppName: "Testbox Backend v1.0",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// 미들웨어 설정
	app.Use(recover.New()) // 패닉 복구
	app.Use(logger.New())  // 로깅
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept",
		AllowCredentials: false,
	}))

	// 라우트 설정
	api.SetupRoutes(app, todoHandler, blogHandler)

	// Graceful Shutdown 설정
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("서버를 안전하게 종료하는 중...")
		app.Shutdown()
	}()

	// 서버 시작
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("🚀 서버 시작: http://localhost%s", addr)
	log.Printf("📝 API 주소: http://localhost%s/api/todos", addr)
	log.Printf("💚 헬스 체크: http://localhost%s/health", addr)

	if err := app.Listen(addr); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}

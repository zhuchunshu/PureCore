package main

import (
	"log"
	"os"

	middleware "purecore/app/Http/Middleware"
	"purecore/core"
	"purecore/routes"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using default ports")
	}

	core.InitLang("lang")

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			res := core.NewResponse(c)
			return res.Error(err.Error(), 500)
		},
	})

	// 全局中间件
	app.Use(middleware.Cors())
	app.Use(middleware.Lang())

	// 注册路由
	router := core.NewRouter(app)
	routes.RegisterAPI(router)

	// 从环境变量读取端口，默认 9002
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "9002"
	}

	log.Printf("PureCore server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

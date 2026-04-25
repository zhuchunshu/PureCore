package cmd

import (
	"log"
	"os"

	middleware "purecore/app/Http/Middleware"
	"purecore/core"
	_ "purecore/database/migrations"
	"purecore/routes"

	"github.com/gofiber/fiber/v3"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the PureCore HTTP server",
	Long:  `Start the PureCore HTTP server on the configured port (default: 9002).`,
	Run:   serveRun,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func serveRun(cmd *cobra.Command, args []string) {
	core.InitLang("lang")

	// Initialize database connection and run auto-migration
	_ = core.DB()
	core.RunMigrations()

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

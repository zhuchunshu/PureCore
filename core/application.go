package core

import (
	"log"

	"github.com/gofiber/fiber/v3"
)

// Application is the main application container that manages the lifecycle
// of all core services: config, database, language, routing, and service providers.
// It provides a clean bootstrap process for the entire framework.
type Application struct {
	config   *Config
	app      *fiber.App
	router   *Router
	registry *ServiceProviderRegistry
	langDir  string
}

// NewApplication creates a new Application instance with all defaults.
// Use the builder methods to configure before calling Boot().
func NewApplication() *Application {
	return &Application{
		config:   GetConfig(),
		registry: NewServiceProviderRegistry(),
		langDir:  "lang",
	}
}

// Config returns the application configuration
func (a *Application) Config() *Config {
	return a.config
}

// App returns the underlying Fiber app instance
func (a *Application) App() *fiber.App {
	return a.app
}

// Router returns the router instance
func (a *Application) Router() *Router {
	return a.router
}

// SetLangDir sets the directory for language files
func (a *Application) SetLangDir(dir string) *Application {
	a.langDir = dir
	return a
}

// AddProviders registers service providers with the application.
// Providers will be registered and booted in order during Boot().
func (a *Application) AddProviders(providers ...ServiceProvider) *Application {
	a.registry.Add(providers...)
	return a
}

// Boot initializes all application services: config, language, database,
// middleware, service providers, and starts the HTTP server.
func (a *Application) Boot() error {
	// Initialize language
	InitLang(a.langDir)

	// Initialize database
	_ = DB()
	RunMigrations()

	// Create Fiber app with error handler
	a.app = fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			res := NewResponse(c)
			return res.Error(err.Error(), 500)
		},
	})

	// Create router
	a.router = NewRouter(a.app)

	// Register all service providers (which register routes)
	if err := a.registry.RegisterAll(a.router); err != nil {
		return err
	}

	// Boot all service providers
	if err := a.registry.BootAll(); err != nil {
		return err
	}

	return nil
}

// Run starts the HTTP server on the configured port
func (a *Application) Run() error {
	port := a.config.AppPort()
	log.Printf("PureCore server starting on port %s", port)
	return a.app.Listen(":" + port)
}

// RunWithMiddleware is a convenience method that applies global middleware
// before registering providers and starting the server.
// Use this for middleware that should run on every request.
func (a *Application) RunWithMiddleware(middlewares ...fiber.Handler) error {
	if err := a.Boot(); err != nil {
		return err
	}

	// Apply global middleware
	for _, m := range middlewares {
		a.app.Use(m)
	}

	// Start server
	port := a.config.AppPort()
	if port == "" {
		port = "9002"
	}

	// Check for PORT environment variable (common in hosting environments like Heroku, Railway, etc.)
	// The PORT variable is a PaaS standard that may differ from BACKEND_PORT
	if envPort := GetConfig().String("PORT"); envPort != "" {
		port = envPort
	}

	log.Printf("PureCore server starting on port %s", port)
	return a.app.Listen(":" + port)
}

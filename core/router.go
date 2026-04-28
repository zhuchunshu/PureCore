package core

import "github.com/gofiber/fiber/v3"

// Router provides a fluent, Laravel-style route builder with support for
// prefix groups, middleware chains, and named middleware resolution.
// Named middleware allows registering reusable middleware by name and
// referencing them in route definitions for cleaner, more maintainable code.
type Router struct {
	app             *fiber.App
	prefix          string
	middlewares     []fiber.Handler
	namedMiddleware map[string]fiber.Handler
}

var App *fiber.App

// NewRouter creates a new Router instance.
// If no named middleware map is provided, initializes an empty one.
func NewRouter(app *fiber.App) *Router {
	App = app
	return &Router{
		app:             app,
		namedMiddleware: make(map[string]fiber.Handler),
	}
}

// RegisterNamedMiddleware registers a named middleware that can be referenced
// by name in route definitions using MiddlewareByName().
func (r *Router) RegisterNamedMiddleware(name string, handler fiber.Handler) *Router {
	r.namedMiddleware[name] = handler
	return r
}

// RegisterNamedMiddlewares registers multiple named middleware handlers at once.
func (r *Router) RegisterNamedMiddlewares(middlewares map[string]fiber.Handler) *Router {
	for name, handler := range middlewares {
		r.namedMiddleware[name] = handler
	}
	return r
}

// MiddlewareByName adds middleware by its registered name.
// This allows routes to reference pre-registered middleware without
// importing the middleware package directly.
func (r *Router) MiddlewareByName(names ...string) *Router {
	var handlers []fiber.Handler
	for _, name := range names {
		if handler, ok := r.namedMiddleware[name]; ok {
			handlers = append(handlers, handler)
		}
	}
	newMiddlewares := append(r.middlewares, handlers...)
	return &Router{app: r.app, prefix: r.prefix, middlewares: newMiddlewares, namedMiddleware: r.namedMiddleware}
}

// GetNamedMiddleware returns the named middleware map for external access.
// Useful for service providers that want to register middleware centrally.
func (r *Router) GetNamedMiddleware() map[string]fiber.Handler {
	return r.namedMiddleware
}

// Prefix adds a path prefix to the router group, supporting Laravel-style
// route grouping: Route::prefix("/api/v1")->middleware(...)->group(...)
func (r *Router) Prefix(prefix string) *Router {
	return &Router{
		app:             r.app,
		prefix:          r.prefix + prefix,
		middlewares:     r.middlewares,
		namedMiddleware: r.namedMiddleware,
	}
}

// Middleware appends one or more middleware handlers to the router group.
func (r *Router) Middleware(handlers ...fiber.Handler) *Router {
	newMiddlewares := make([]fiber.Handler, len(r.middlewares)+len(handlers))
	copy(newMiddlewares, r.middlewares)
	copy(newMiddlewares[len(r.middlewares):], handlers)
	return &Router{
		app:             r.app,
		prefix:          r.prefix,
		middlewares:     newMiddlewares,
		namedMiddleware: r.namedMiddleware,
	}
}

// Group executes a callback with the current router as the group root,
// enabling nested route definitions.
func (r *Router) Group(fn func(r *Router)) {
	fn(r)
}

// addRoute registers a route with the specified HTTP method, path, and handler.
// Automatically applies all accumulated middleware to the route.
func (r *Router) addRoute(method func(path string, handler any, middleware ...any) fiber.Router, path string, handler fiber.Handler) {
	handlers := append(r.middlewares, handler)
	if len(handlers) == 0 {
		return
	}
	// Convert fiber.Handler slice to any slice for fiber v3 API
	anyHandlers := make([]any, len(handlers))
	for i, h := range handlers {
		anyHandlers[i] = h
	}
	method(r.prefix+path, anyHandlers[0], anyHandlers[1:]...)
}

// Get registers a GET route at the given path.
func (r *Router) Get(path string, handler fiber.Handler) {
	r.addRoute(r.app.Get, path, handler)
}

// Post registers a POST route at the given path.
func (r *Router) Post(path string, handler fiber.Handler) {
	r.addRoute(r.app.Post, path, handler)
}

// Put registers a PUT route at the given path.
func (r *Router) Put(path string, handler fiber.Handler) {
	r.addRoute(r.app.Put, path, handler)
}

// Delete registers a DELETE route at the given path.
func (r *Router) Delete(path string, handler fiber.Handler) {
	r.addRoute(r.app.Delete, path, handler)
}

// Patch registers a PATCH route at the given path.
func (r *Router) Patch(path string, handler fiber.Handler) {
	r.addRoute(r.app.Patch, path, handler)
}

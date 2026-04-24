package core

import "github.com/gofiber/fiber/v3"

type Router struct {
	app         *fiber.App
	prefix      string
	middlewares []fiber.Handler
}

var App *fiber.App

func NewRouter(app *fiber.App) *Router {
	App = app
	return &Router{app: app}
}

// 支持分组，类似 Route::prefix()->middleware()->group()
func (r *Router) Prefix(prefix string) *Router {
	return &Router{app: r.app, prefix: r.prefix + prefix, middlewares: r.middlewares}
}

func (r *Router) Middleware(handlers ...fiber.Handler) *Router {
	newMiddlewares := append(r.middlewares, handlers...)
	return &Router{app: r.app, prefix: r.prefix, middlewares: newMiddlewares}
}

func (r *Router) Group(fn func(r *Router)) {
	fn(r)
}

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

func (r *Router) Get(path string, handler fiber.Handler) {
	r.addRoute(r.app.Get, path, handler)
}

func (r *Router) Post(path string, handler fiber.Handler) {
	r.addRoute(r.app.Post, path, handler)
}

func (r *Router) Put(path string, handler fiber.Handler) {
	r.addRoute(r.app.Put, path, handler)
}

func (r *Router) Delete(path string, handler fiber.Handler) {
	r.addRoute(r.app.Delete, path, handler)
}

func (r *Router) Patch(path string, handler fiber.Handler) {
	r.addRoute(r.app.Patch, path, handler)
}

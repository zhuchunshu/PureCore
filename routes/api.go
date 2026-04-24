package routes

import (
	controllers "purecore/app/Http/Controllers"
	middleware "purecore/app/Http/Middleware"
	"purecore/core"
)

func RegisterAPI(r *core.Router) {
	userCtrl := &controllers.UserController{}
	sysCtrl := &controllers.SystemController{}

	// 公开路由
	r.Prefix("/api/v1").Group(func(r *core.Router) {
		r.Get("/ping", core.H(func(req *core.Request, res *core.Response) error {
			return res.Success("pong")
		}))
		r.Get("/system/info", core.H(sysCtrl.Info))
	})

	// 需要鉴权的路由组
	r.Prefix("/api/v1").Middleware(middleware.Auth()).Group(func(r *core.Router) {
		r.Get("/users", core.H(userCtrl.Index))
		r.Post("/users", core.H(userCtrl.Store))
		r.Get("/users/:id", core.H(userCtrl.Show))
	})
}

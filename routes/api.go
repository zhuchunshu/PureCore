package routes

import (
	controllers "purecore/app/Http/Controllers"
	middleware "purecore/app/Http/Middleware"
	"purecore/core"
)

func RegisterAPI(r *core.Router) {
	userCtrl := &controllers.UserController{}
	userAuthCtrl := &controllers.UserAuthController{}
	sysCtrl := &controllers.SystemController{}
	optionCtrl := &controllers.OptionController{}

	// 公开路由
	r.Prefix("/api/v1").Group(func(r *core.Router) {
		r.Get("/ping", core.H(func(req *core.Request, res *core.Response) error {
			return res.Success("pong")
		}))
		r.Get("/system/info", core.H(sysCtrl.Info))
		// 公共用户认证路由
		r.Post("/auth/register", core.H(userAuthCtrl.Register))
		r.Post("/auth/login", core.H(userAuthCtrl.Login))
		r.Post("/auth/refresh", core.H(userAuthCtrl.Refresh))
	})

	// 管理员路由（动态前缀，从 .env 读取）
	adminCtrl := &controllers.AdminAuthController{}
	adminPrefix := controllers.GetAdminRoutePrefix()

	// 公开的管理员路由（无需鉴权）
	r.Prefix(adminPrefix).Group(func(r *core.Router) {
		r.Get("/auth/check", core.H(adminCtrl.CheckAdminExists))
		r.Post("/auth/login", core.H(adminCtrl.Login))
		r.Post("/auth/register", core.H(adminCtrl.CreateAdmin))
		r.Post("/auth/refresh", core.H(adminCtrl.Refresh))
		r.Get("/options", core.H(optionCtrl.GetAll))
	})

	// 需要管理员鉴权的路由
	r.Prefix(adminPrefix).Middleware(middleware.AdminAuth()).Group(func(r *core.Router) {
		r.Get("/auth/profile", core.H(adminCtrl.Profile))
		r.Post("/auth/change-password", core.H(adminCtrl.ChangePassword))
		r.Post("/options", core.H(optionCtrl.Set))
	})

	// 需要鉴权的路由组
	r.Prefix("/api/v1").Middleware(middleware.Auth()).Group(func(r *core.Router) {
		r.Get("/auth/profile", core.H(userAuthCtrl.Profile))
		r.Get("/users", core.H(userCtrl.Index))
		r.Post("/users", core.H(userCtrl.Store))
		r.Get("/users/:id", core.H(userCtrl.Show))
	})
}

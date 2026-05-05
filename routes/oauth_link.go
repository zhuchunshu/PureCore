package routes

import (
	controllers "purecore/app/Http/Controllers"
	"purecore/core"
)

// RegisterOAuthLinkRoutes registers routes for OAuth account linking.
// These endpoints handle the case where an OAuth identity is not yet linked
// to any PureCore user, providing register+link and login+link flows.
func RegisterOAuthLinkRoutes(router *core.Router) {
	ctrl := &controllers.OAuthLinkController{}
	router.Post("/api/v1/oauth/:provider/link/register", core.H(ctrl.LinkRegister))
	router.Post("/api/v1/oauth/:provider/link/login", core.H(ctrl.LinkLogin))
}

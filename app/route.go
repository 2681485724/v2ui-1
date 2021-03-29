package app

import (
	"github.com/gin-gonic/gin"
)

//注册路由
func RegisterRoutes(g *gin.Engine) {
	new(DataBaseController).RegisterRoutes(g)
	new(StatusController).RegisterRoutes(g)
	new(V2RayController).RegisterRoutes(g)
	new(TemplateController).RegisterRoutes(g)
	new(ConfigController).RegisterRoutes(g)
	new(URLController).RegisterRoutes(g)
	new(SUBController).RegisterRoutes(g)
}

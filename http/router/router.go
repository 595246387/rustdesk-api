package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/controller/web"
	"net/http"
	"strings"
)

// adminCacheControl 后台前端的缓存策略。
//
// index.html 不缓存，保证发版后浏览器立即拿到新的入口 chunk；
// static/ 下的文件名带内容 hash，可以长缓存。
// 由后端设置而不是依赖 nginx，避免中间多一层反代时响应头被覆盖。
func adminCacheControl(c *gin.Context) {
	path := c.Request.URL.Path
	switch {
	case strings.HasPrefix(path, "/_admin/static/"):
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
	case path == "/_admin/" || path == "/_admin/index.html":
		c.Header("Cache-Control", "no-cache, must-revalidate")
	}
	c.Next()
}

func WebInit(g *gin.Engine) {
	g.Use(adminCacheControl)

	i := &web.Index{}
	g.GET("/", i.Index)

	if global.Config.App.WebClient == 1 {
		g.GET("/webclient-config/index.js", i.ConfigJs)
	}

	if global.Config.App.WebClient == 1 {
		g.StaticFS("/webclient", http.Dir(global.Config.Gin.ResourcesPath+"/web"))
		g.StaticFS("/webclient2", http.Dir(global.Config.Gin.ResourcesPath+"/web2"))
	}
	g.StaticFS("/_admin", http.Dir(global.Config.Gin.ResourcesPath+"/admin"))
}

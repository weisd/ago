package routers

import (
	"github.com/weisd/ago/modules/log"
	"github.com/weisd/ago/modules/middleware"
)

func Home(ctx *middleware.Context) {
	log.Trace("首页")
	ctx.Session.Set("uid", "dada")
	ctx.Data["Name"] = "weisd"
	ctx.HTML(200, "site/home", ctx.Data)
}

func Test(ctx *middleware.Context) {
	log.Info("info : %s", ctx.Session.Get("uid"))
	log.Info("csrf : %s", ctx.Data["CsrfToken"])
}

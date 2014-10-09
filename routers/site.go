package routers

import (
	"github.com/Unknwon/macaron"
)

func Home(ctx *macaron.Context) {
	ctx.Data["Name"] = "weisd"
	ctx.HTML(200, "site/home", ctx.Data)
}

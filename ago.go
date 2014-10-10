package main

import (
	"github.com/Unknwon/macaron"
	"github.com/macaron-contrib/cache"
	"github.com/macaron-contrib/captcha"
	"github.com/macaron-contrib/csrf"
	"github.com/macaron-contrib/session"
	"github.com/macaron-contrib/toolbox"
	"github.com/weisd/ago/models"
	"github.com/weisd/ago/modules/middleware"
	"github.com/weisd/ago/modules/setting"
	"github.com/weisd/ago/routers"
	"runtime"
)

const APP_VER = "0.0.1"

func init() {
	// goroutine 使用核心数
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	routers.GlobalInit()

	m := newMacaron()

	m.Get("/", routers.Home)
	m.Get("/test", routers.Test)

	listenAddr := ":" + setting.HttpPort

	m.RunOnAddr(listenAddr)
}

func newMacaron() *macaron.Macaron {
	m := macaron.Classic()
	// 必须要加入render @todo 映射funcs
	m.Use(macaron.Renderer())

	m.Use(cache.Cacher(cache.Options{
		Adapter:  setting.CacheAdapter,
		Interval: setting.CacheInternal,
		Conn:     setting.CacheConn,
	}))
	//  captcha要放在cache 后面
	m.Use(captcha.Captchaer(captcha.Options{
		SubURL: setting.AppSubUrl,
	}))

	m.Use(session.Sessioner(session.Options{
		Provider: setting.SessionProvider,
		Config:   *setting.SessionConfig,
	}))

	m.Use(csrf.Generate(csrf.Options{
		Secret:     setting.SecretKey,
		SetCookie:  true,
		Header:     "X-Csrf-Token",
		CookiePath: setting.AppSubUrl,
	}))

	m.Use(toolbox.Toolboxer(m, toolbox.Options{
		HealthCheckFuncs: []*toolbox.HealthCheckFuncDesc{
			&toolbox.HealthCheckFuncDesc{
				Desc: "Database connection",
				Func: models.Ping,
			},
		},
	}))

	m.Use(middleware.Contexter())

	return m
}

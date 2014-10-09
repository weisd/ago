package main

import (
	"github.com/Unknwon/macaron"
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
	m := macaron.Classic()
	// 必须要加入render
	m.Use(macaron.Renderer())

	m.Get("/", routers.Home)

	listenAddr := ":" + setting.HttpPort

	m.RunOnAddr(listenAddr)
}

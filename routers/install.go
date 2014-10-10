package routers

import (
	"github.com/Unknwon/macaron"
	"github.com/weisd/ago/models"
	"github.com/weisd/ago/modules/log"
	"github.com/weisd/ago/modules/mailer"
	"github.com/weisd/ago/modules/setting"
	"strings"
)

func checkRunMode() {
	switch setting.Cfg.MustValue("", "RUN_MODE") {
	case "prod":
		macaron.Env = macaron.PROD
		setting.ProdMode = true
	case "test":
		macaron.Env = macaron.TEST
	}
	log.Info("Run Mode: %s", strings.Title(macaron.Env))
}

func GlobalInit() {
	setting.NewConfigContext()

	setting.NewServices()
	checkRunMode()

	log.Info("Custom path: %s", setting.CustomPath)
	log.Info("Log path: %s", setting.LogRootPath)
	// 邮件监听
	mailer.NewMailerContext()

	// mysql
	models.LoadModelsConfig()
	if err := models.NewEngine(); err != nil {
		log.Fatal(4, "Fail to initialize ORM engine: %v", err)
	}
}

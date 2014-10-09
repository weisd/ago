package setting

import (
	"fmt"
	"github.com/Unknwon/com"
	"github.com/Unknwon/goconfig"
	"github.com/Unknwon/macaron"
)

const (
	CFG_PATH        = "conf/app.ini"
	CFG_CUSTOM_PATH = "conf/custom.ini"
)

var (
	Cfg      *goconfig.ConfigFile
	HttpPort string
)

func init() {
	var err error
	Cfg, err = goconfig.LoadConfigFile(CFG_PATH)
	if err != nil {
		panic(fmt.Errorf("fail to load config file '%s': %v", CFG_PATH, err))
	}

	if com.IsFile(CFG_CUSTOM_PATH) {
		if err = Cfg.AppendFiles(CFG_CUSTOM_PATH); err != nil {
			panic(fmt.Errorf("fail to append config file '%s' : %v", CFG_CUSTOM_PATH, err))
		}
	}

	// 生产环境
	if Cfg.MustValue("app", "run_mode", "dev") == "prod" {
		macaron.Env = macaron.PROD
	}

	HttpPort = Cfg.MustValue("app", "http_port", "8888")
}

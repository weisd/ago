package setting

import (
	"fmt"
	"github.com/Unknwon/com"
	"github.com/Unknwon/goconfig"
	"github.com/macaron-contrib/session"
	// "github.com/Unknwon/macaron"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/weisd/ago/modules/log"
)

type Scheme string

const (
	HTTP  Scheme = "http"
	HTTPS Scheme = "https"
)

// Mailer represents mail service.
type Mailer struct {
	Name         string
	Host         string
	From         string
	User, Passwd string
}

var (
	// App settings.
	AppVer    string
	AppName   string
	AppUrl    string
	AppSubUrl string

	// Server settings.
	Protocol           Scheme
	Domain             string
	HttpAddr, HttpPort string
	SshPort            int
	OfflineMode        bool
	DisableRouterLog   bool
	CertFile, KeyFile  string
	StaticRootPath     string
	EnableGzip         bool

	Cfg          *goconfig.ConfigFile
	ConfRootPath string
	CustomPath   string // Custom directory path.
	ProdMode     bool

	// Log settings.
	LogRootPath string
	LogModes    []string
	LogConfigs  []string

	MailService *Mailer

	// Cache settings.
	CacheAdapter  string
	CacheInternal int
	CacheConn     string

	EnableRedis    bool
	EnableMemcache bool

	// Session settings.
	SessionProvider string
	SessionConfig   *session.Config

	SecretKey            string
	LogInRememberDays    int
	CookieUserName       string
	CookieRememberName   string
	ReverseProxyAuthUser string
)

var logLevels = map[string]string{
	"Trace":    "0",
	"Debug":    "1",
	"Info":     "2",
	"Warn":     "3",
	"Error":    "4",
	"Critical": "5",
}

func ExecPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	p, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	return p, nil
}

// WorkDir returns absolute path of work directory.
func WorkDir() (string, error) {
	execPath, err := ExecPath()
	return path.Dir(strings.Replace(execPath, "\\", "/", -1)), err
}

// NewConfigContext initializes configuration context.
// NOTE: do not print any log except error.
func NewConfigContext() {
	workDir, err := WorkDir()
	if err != nil {
		log.Fatal(4, "Fail to get work directory: %v", err)
	}
	ConfRootPath = path.Join(workDir, "conf")

	Cfg, err = goconfig.LoadConfigFile(path.Join(workDir, "conf/app.ini"))
	if err != nil {
		log.Fatal(4, "Fail to parse 'conf/app.ini': %v", err)
	}

	CustomPath = os.Getenv("GOGS_CUSTOM")
	if len(CustomPath) == 0 {
		CustomPath = path.Join(workDir, "custom")
	}

	cfgPath := path.Join(CustomPath, "conf/app.ini")
	if com.IsFile(cfgPath) {
		if err = Cfg.AppendFiles(cfgPath); err != nil {
			log.Fatal(4, "Fail to load custom 'conf/app.ini': %v", err)
		}
	} else {
		log.Warn("No custom 'conf/app.ini' found, please go to '/install'")
	}

	AppName = Cfg.MustValue("", "APP_NAME", "Gogs: Go Git Service")
	AppUrl = Cfg.MustValue("server", "ROOT_URL", "http://localhost:3000/")
	if AppUrl[len(AppUrl)-1] != '/' {
		AppUrl += "/"
	}

	// Check if has app suburl.
	url, err := url.Parse(AppUrl)
	if err != nil {
		log.Fatal(4, "Invalid ROOT_URL(%s): %s", AppUrl, err)
	}
	AppSubUrl = strings.TrimSuffix(url.Path, "/")

	Protocol = HTTP
	if Cfg.MustValue("server", "PROTOCOL") == "https" {
		Protocol = HTTPS
		CertFile = Cfg.MustValue("server", "CERT_FILE")
		KeyFile = Cfg.MustValue("server", "KEY_FILE")
	}
	Domain = Cfg.MustValue("server", "DOMAIN", "localhost")
	HttpAddr = Cfg.MustValue("server", "HTTP_ADDR", "0.0.0.0")
	HttpPort = Cfg.MustValue("server", "HTTP_PORT", "3000")
	SshPort = Cfg.MustInt("server", "SSH_PORT", 22)
	OfflineMode = Cfg.MustBool("server", "OFFLINE_MODE")
	DisableRouterLog = Cfg.MustBool("server", "DISABLE_ROUTER_LOG")
	StaticRootPath = Cfg.MustValue("server", "STATIC_ROOT_PATH", workDir)
	LogRootPath = Cfg.MustValue("log", "ROOT_PATH", path.Join(workDir, "log"))
	EnableGzip = Cfg.MustBool("server", "ENABLE_GZIP")

	SecretKey = Cfg.MustValue("security", "SECRET_KEY")
	LogInRememberDays = Cfg.MustInt("security", "LOGIN_REMEMBER_DAYS")
	CookieUserName = Cfg.MustValue("security", "COOKIE_USERNAME")
	CookieRememberName = Cfg.MustValue("security", "COOKIE_REMEMBER_NAME")
	ReverseProxyAuthUser = Cfg.MustValue("security", "REVERSE_PROXY_AUTHENTICATION_USER", "X-WEBAUTH-USER")

}

func newLogService() {
	// log.Info("%s %s", AppName, AppVer)

	// Get and check log mode.
	LogModes = strings.Split(Cfg.MustValue("log", "MODE", "console"), ",")
	LogConfigs = make([]string, len(LogModes))
	for i, mode := range LogModes {
		mode = strings.TrimSpace(mode)
		modeSec := "log." + mode
		if _, err := Cfg.GetSection(modeSec); err != nil {
			log.Fatal(4, "Unknown log mode: %s", mode)
		}

		// Log level.
		levelName := Cfg.MustValueRange("log."+mode, "LEVEL", "Trace",
			[]string{"Trace", "Debug", "Info", "Warn", "Error", "Critical"})
		level, ok := logLevels[levelName]
		if !ok {
			log.Fatal(4, "Unknown log level: %s", levelName)
		}

		// Generate log configuration.
		switch mode {
		case "console":
			LogConfigs[i] = fmt.Sprintf(`{"level":%s}`, level)
		case "file":
			logPath := Cfg.MustValue(modeSec, "FILE_NAME", path.Join(LogRootPath, "gogs.log"))
			os.MkdirAll(path.Dir(logPath), os.ModePerm)
			LogConfigs[i] = fmt.Sprintf(
				`{"level":%s,"filename":"%s","rotate":%v,"maxlines":%d,"maxsize":%d,"daily":%v,"maxdays":%d}`, level,
				logPath,
				Cfg.MustBool(modeSec, "LOG_ROTATE", true),
				Cfg.MustInt(modeSec, "MAX_LINES", 1000000),
				1<<uint(Cfg.MustInt(modeSec, "MAX_SIZE_SHIFT", 28)),
				Cfg.MustBool(modeSec, "DAILY_ROTATE", true),
				Cfg.MustInt(modeSec, "MAX_DAYS", 7))
		case "conn":
			LogConfigs[i] = fmt.Sprintf(`{"level":%s,"reconnectOnMsg":%v,"reconnect":%v,"net":"%s","addr":"%s"}`, level,
				Cfg.MustBool(modeSec, "RECONNECT_ON_MSG"),
				Cfg.MustBool(modeSec, "RECONNECT"),
				Cfg.MustValueRange(modeSec, "PROTOCOL", "tcp", []string{"tcp", "unix", "udp"}),
				Cfg.MustValue(modeSec, "ADDR", ":7020"))
		case "smtp":
			LogConfigs[i] = fmt.Sprintf(`{"level":%s,"username":"%s","password":"%s","host":"%s","sendTos":"%s","subject":"%s"}`, level,
				Cfg.MustValue(modeSec, "USER", "example@example.com"),
				Cfg.MustValue(modeSec, "PASSWD", "******"),
				Cfg.MustValue(modeSec, "HOST", "127.0.0.1:25"),
				Cfg.MustValue(modeSec, "RECEIVERS", "[]"),
				Cfg.MustValue(modeSec, "SUBJECT", "Diagnostic message from serve"))
		case "database":
			LogConfigs[i] = fmt.Sprintf(`{"level":%s,"driver":"%s","conn":"%s"}`, level,
				Cfg.MustValue(modeSec, "DRIVER"),
				Cfg.MustValue(modeSec, "CONN"))
		}

		log.NewLogger(Cfg.MustInt64("log", "BUFFER_LEN", 10000), mode, LogConfigs[i])
		log.Info("Log Mode: %s(%s)", strings.Title(mode), levelName)
	}
}

func newMailService() {
	// Check mailer setting.
	if !Cfg.MustBool("mailer", "ENABLED") {
		return
	}

	MailService = &Mailer{
		Name:   Cfg.MustValue("mailer", "NAME", AppName),
		Host:   Cfg.MustValue("mailer", "HOST"),
		User:   Cfg.MustValue("mailer", "USER"),
		Passwd: Cfg.MustValue("mailer", "PASSWD"),
	}
	MailService.From = Cfg.MustValue("mailer", "FROM", MailService.User)
	log.Info("Mail Service Enabled")
}

func newCacheService() {
	CacheAdapter = Cfg.MustValueRange("cache", "ADAPTER", "memory", []string{"memory", "redis", "memcache"})
	if EnableRedis {
		log.Info("Redis Enabled")
	}
	if EnableMemcache {
		log.Info("Memcache Enabled")
	}

	switch CacheAdapter {
	case "memory":
		CacheInternal = Cfg.MustInt("cache", "INTERVAL", 60)
	case "redis", "memcache":
		CacheConn = strings.Trim(Cfg.MustValue("cache", "HOST"), "\" ")
	default:
		log.Fatal(4, "Unknown cache adapter: %s", CacheAdapter)
	}

	log.Info("Cache Service Enabled")
}

func newSessionService() {
	SessionProvider = Cfg.MustValueRange("session", "PROVIDER", "memory",
		[]string{"memory", "file", "redis", "mysql"})

	SessionConfig = new(session.Config)
	SessionConfig.ProviderConfig = strings.Trim(Cfg.MustValue("session", "PROVIDER_CONFIG"), "\" ")
	SessionConfig.CookieName = Cfg.MustValue("session", "COOKIE_NAME", "i_like_gogits")
	SessionConfig.CookiePath = AppSubUrl
	SessionConfig.Secure = Cfg.MustBool("session", "COOKIE_SECURE")
	SessionConfig.EnableSetCookie = Cfg.MustBool("session", "ENABLE_SET_COOKIE", true)
	SessionConfig.Gclifetime = Cfg.MustInt64("session", "GC_INTERVAL_TIME", 86400)
	SessionConfig.Maxlifetime = Cfg.MustInt64("session", "SESSION_LIFE_TIME", 86400)
	SessionConfig.SessionIDHashFunc = Cfg.MustValueRange("session", "SESSION_ID_HASHFUNC",
		"sha1", []string{"sha1", "sha256", "md5"})
	SessionConfig.SessionIDHashKey = Cfg.MustValue("session", "SESSION_ID_HASHKEY", string(com.RandomCreateBytes(16)))

	if SessionProvider == "file" {
		os.MkdirAll(path.Dir(SessionConfig.ProviderConfig), os.ModePerm)
	}

	log.Info("Session Service Enabled")
}

var Service struct {
	RegisterEmailConfirm   bool
	DisableRegistration    bool
	RequireSignInView      bool
	EnableCacheAvatar      bool
	EnableNotifyMail       bool
	EnableReverseProxyAuth bool
	LdapAuth               bool
	ActiveCodeLives        int
	ResetPwdCodeLives      int
	EnableGitHooks         bool
}

func newService() {
	Service.ActiveCodeLives = Cfg.MustInt("service", "ACTIVE_CODE_LIVE_MINUTES", 180)
	Service.ResetPwdCodeLives = Cfg.MustInt("service", "RESET_PASSWD_CODE_LIVE_MINUTES", 180)
	Service.DisableRegistration = Cfg.MustBool("service", "DISABLE_REGISTRATION")
	Service.RequireSignInView = Cfg.MustBool("service", "REQUIRE_SIGNIN_VIEW")
	Service.EnableCacheAvatar = Cfg.MustBool("service", "ENABLE_CACHE_AVATAR")
	Service.EnableReverseProxyAuth = Cfg.MustBool("service", "ENABLE_REVERSE_PROXY_AUTHENTICATION")
	Service.EnableGitHooks = Cfg.MustBool("service", "ENABLE_GIT_HOOKS")
}

func NewServices() {
	newService()
	newLogService()
	newMailService()
	newCacheService()
	newSessionService()
}

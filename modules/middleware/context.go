package middleware

import (
	"github.com/Unknwon/macaron"
	"github.com/macaron-contrib/cache"
	"github.com/macaron-contrib/csrf"
	"github.com/macaron-contrib/session"
	"github.com/weisd/ago/models"
	"github.com/weisd/ago/modules/setting"
	"html/template"
	"strings"
	"time"
)

type Context struct {
	*macaron.Context
	Cache   cache.Cache
	csrf    csrf.CSRF
	Flash   *session.Flash
	Session session.Store

	User *models.User

	IsSigned bool
}

// Query querys form parameter.
func (ctx *Context) Query(name string) string {
	ctx.Req.ParseForm()
	return ctx.Req.Form.Get(name)
}

// Contexter initializes a classic context for a request.
func Contexter() macaron.Handler {
	return func(c *macaron.Context, cah cache.Cache, sess session.Store, f *session.Flash, x csrf.CSRF) {
		ctx := &Context{
			Context: c,
			Cache:   cah,
			csrf:    x,
			Flash:   f,
			Session: sess,
		}
		// Compute current URL for real-time change language.
		link := setting.AppSubUrl + ctx.Req.RequestURI
		i := strings.Index(link, "?")
		if i > -1 {
			link = link[:i]
		}
		ctx.Data["Link"] = link

		ctx.Data["PageStartTime"] = time.Now()

		// // Get user from session if logined.
		// ctx.User = auth.SignedInUser(ctx.Req.Header, ctx.Session)
		// if ctx.User != nil {
		// 	ctx.IsSigned = true
		// 	ctx.Data["IsSigned"] = ctx.IsSigned
		// 	ctx.Data["SignedUser"] = ctx.User
		// 	ctx.Data["IsAdmin"] = ctx.User.IsAdmin
		// }

		// // If request sends files, parse them here otherwise the Query() can't be parsed and the CsrfToken will be invalid.
		// if ctx.Req.Method == "POST" && strings.Contains(ctx.Req.Header.Get("Content-Type"), "multipart/form-data") {
		// 	if err := ctx.Req.ParseMultipartForm(setting.AttachmentMaxSize << 20); err != nil && !strings.Contains(err.Error(), "EOF") { // 32MB max size
		// 		ctx.Handle(500, "ParseMultipartForm", err)
		// 		return
		// 	}
		// }

		ctx.Data["CsrfToken"] = x.GetToken()
		ctx.Data["CsrfTokenHtml"] = template.HTML(`<input type="hidden" name="_csrf" value="` + x.GetToken() + `">`)

		c.Map(ctx)
	}
}

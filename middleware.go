package kocha

import (
	"strconv"

	"github.com/naoina/kocha/util"
)

// Middleware is the interface that middleware.
type Middleware interface {
	Before(app *Application, c *Controller)
	After(app *Application, c *Controller)
}

var (
	// Default middlewares.
	DefaultMiddlewares = []Middleware{
		&ResponseContentTypeMiddleware{},
	}
)

// Middleware that set Content-Type header.
type ResponseContentTypeMiddleware struct{}

func (m *ResponseContentTypeMiddleware) Before(app *Application, c *Controller) {
	// do nothing
}

func (m *ResponseContentTypeMiddleware) After(app *Application, c *Controller) {
	res := c.Response
	res.Header().Set("Content-Type", res.ContentType)
}

// Session processing middleware.
type SessionMiddleware struct{}

func (m *SessionMiddleware) Before(app *Application, c *Controller) {
	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case ErrSession:
				Log.Error("%v", err)
			case ErrSessionExpected:
				Log.Info("%v", err)
			default:
				panic(err)
			}
			c.Session = make(Session)
		}
	}()
	cookie, err := c.Request.Cookie(app.Config.Session.Name)
	if err != nil {
		panic(NewErrSessionExpected("new session"))
	}
	sess := app.Config.Session.Store.Load(cookie.Value)
	expiresStr, ok := sess[SessionExpiresKey]
	if !ok {
		panic(NewErrSession("expires value not found"))
	}
	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		panic(NewErrSession(err.Error()))
	}
	if expires < util.Now().Unix() {
		panic(NewErrSessionExpected("session has been expired"))
	}
	c.Session = sess
}

func (m *SessionMiddleware) After(app *Application, c *Controller) {
	expires, _ := expiresFromDuration(app.Config.Session.SessionExpires)
	c.Session[SessionExpiresKey] = strconv.FormatInt(expires.Unix(), 10)
	cookie := newSessionCookie(app, c)
	cookie.Value = app.Config.Session.Store.Save(c.Session)
	c.Response.SetCookie(cookie)
}

// Request logging middleware.
type RequestLoggingMiddleware struct{}

func (m *RequestLoggingMiddleware) Before(app *Application, c *Controller) {
	// do nothing.
}

func (m *RequestLoggingMiddleware) After(app *Application, c *Controller) {
	Log.Info(`"%v %v %v" %v`, c.Request.Method, c.Request.RequestURI, c.Request.Proto, c.Response.StatusCode)
}

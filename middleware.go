package kocha

import (
	"bytes"
	"strconv"

	"github.com/naoina/kocha/log"
	"github.com/naoina/kocha/util"
	"github.com/ugorji/go/codec"
)

// Middleware is the interface that middleware.
type Middleware interface {
	Before(app *Application, c *Context)
	After(app *Application, c *Context)
}

// Session processing middleware.
type SessionMiddleware struct{}

func (m *SessionMiddleware) Before(app *Application, c *Context) {
	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case ErrSession:
				app.Logger.Error(err)
			case ErrSessionExpected:
				app.Logger.Info(err)
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

func (m *SessionMiddleware) After(app *Application, c *Context) {
	expires, _ := expiresFromDuration(app.Config.Session.SessionExpires)
	c.Session[SessionExpiresKey] = strconv.FormatInt(expires.Unix(), 10)
	cookie := newSessionCookie(app, c)
	cookie.Value = app.Config.Session.Store.Save(c.Session)
	c.Response.SetCookie(cookie)
}

// Flash messages processing middleware.
type FlashMiddleware struct{}

func (m *FlashMiddleware) Before(app *Application, c *Context) {
	if c.Session == nil {
		app.Logger.Error("FlashMiddleware hasn't been added after SessionMiddleware; it cannot be used")
		return
	}
	c.Flash = Flash{}
	if flash := c.Session["_flash"]; flash != "" {
		if err := codec.NewDecoderBytes([]byte(flash), codecHandler).Decode(&c.Flash); err != nil {
			// make a new Flash instance because there is a possibility that
			// garbage data is set to c.Flash by in-place decoding of Decode().
			c.Flash = Flash{}
			app.Logger.Errorf("kocha: flash: unexpected error in decode process: %v", err)
		}
	}
}

func (m *FlashMiddleware) After(app *Application, c *Context) {
	if c.Session == nil {
		return
	}
	if c.Flash.deleteLoaded(); c.Flash.Len() == 0 {
		delete(c.Session, "_flash")
		return
	}
	var buf bytes.Buffer
	if err := codec.NewEncoder(&buf, codecHandler).Encode(c.Flash); err != nil {
		app.Logger.Errorf("kocha: flash: unexpected error in encode process: %v", err)
		return
	}
	c.Session["_flash"] = buf.String()
}

// Request logging middleware.
type RequestLoggingMiddleware struct{}

func (m *RequestLoggingMiddleware) Before(app *Application, c *Context) {
	// do nothing.
}

func (m *RequestLoggingMiddleware) After(app *Application, c *Context) {
	app.Logger.With(log.Fields{
		"method":   c.Request.Method,
		"uri":      c.Request.RequestURI,
		"protocol": c.Request.Proto,
		"status":   c.Response.StatusCode,
	}).Info()
}

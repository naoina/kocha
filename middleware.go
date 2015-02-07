package kocha

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/naoina/kocha/log"
	"github.com/naoina/kocha/util"
	"github.com/ugorji/go/codec"
)

// Middleware is the interface that middleware.
type Middleware interface {
	Before(app *Application, c *Context) error
	After(app *Application, c *Context) error
}

// Session processing middleware.
type SessionMiddleware struct{}

func (m *SessionMiddleware) Before(app *Application, c *Context) (err error) {
	defer func() {
		switch err.(type) {
		case ErrSession:
			app.Logger.Info(err)
		default:
			app.Logger.Error(err)
		}
		if c.Session == nil {
			c.Session = make(Session)
		}
		err = nil
	}()
	cookie, err := c.Request.Cookie(app.Config.Session.Name)
	if err != nil {
		return NewErrSession("new session")
	}
	sess, err := app.Config.Session.Store.Load(cookie.Value)
	if err != nil {
		return err
	}
	expiresStr, ok := sess[SessionExpiresKey]
	if !ok {
		return fmt.Errorf("expires value not found")
	}
	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		return err
	}
	if expires < util.Now().Unix() {
		return NewErrSession("session has been expired")
	}
	c.Session = sess
	return nil
}

func (m *SessionMiddleware) After(app *Application, c *Context) (err error) {
	expires, _ := expiresFromDuration(app.Config.Session.SessionExpires)
	c.Session[SessionExpiresKey] = strconv.FormatInt(expires.Unix(), 10)
	cookie := newSessionCookie(app, c)
	cookie.Value, err = app.Config.Session.Store.Save(c.Session)
	if err != nil {
		return err
	}
	c.Response.SetCookie(cookie)
	return nil
}

// Flash messages processing middleware.
type FlashMiddleware struct{}

func (m *FlashMiddleware) Before(app *Application, c *Context) error {
	if c.Session == nil {
		app.Logger.Error("kocha: FlashMiddleware hasn't been added after SessionMiddleware; it cannot be used")
		return nil
	}
	c.Flash = Flash{}
	if flash := c.Session["_flash"]; flash != "" {
		if err := codec.NewDecoderBytes([]byte(flash), codecHandler).Decode(&c.Flash); err != nil {
			// make a new Flash instance because there is a possibility that
			// garbage data is set to c.Flash by in-place decoding of Decode().
			c.Flash = Flash{}
			return fmt.Errorf("kocha: flash: unexpected error in decode process: %v", err)
		}
	}
	return nil
}

func (m *FlashMiddleware) After(app *Application, c *Context) error {
	if c.Session == nil {
		return nil
	}
	if c.Flash.deleteLoaded(); c.Flash.Len() == 0 {
		delete(c.Session, "_flash")
		return nil
	}
	var buf bytes.Buffer
	if err := codec.NewEncoder(&buf, codecHandler).Encode(c.Flash); err != nil {
		return fmt.Errorf("kocha: flash: unexpected error in encode process: %v", err)
	}
	c.Session["_flash"] = buf.String()
	return nil
}

// Request logging middleware.
type RequestLoggingMiddleware struct{}

func (m *RequestLoggingMiddleware) Before(app *Application, c *Context) error {
	return nil
}

func (m *RequestLoggingMiddleware) After(app *Application, c *Context) error {
	app.Logger.With(log.Fields{
		"method":   c.Request.Method,
		"uri":      c.Request.RequestURI,
		"protocol": c.Request.Proto,
		"status":   c.Response.StatusCode,
	}).Info()
	return nil
}

package kocha

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/naoina/kocha/log"
	"github.com/naoina/kocha/util"
	"github.com/ugorji/go/codec"
)

// Middleware is the interface that middleware.
type Middleware interface {
	Process(app *Application, c *Context, next func() error) error
}

// Validator is the interface to validate the middleware.
type Validator interface {
	// Validate validates the middleware.
	// Validate will be called in initializing the application.
	Validate() error
}

// PanicRecoverMiddleware is a middleware to recover a panic where occurred in request sequence.
type PanicRecoverMiddleware struct{}

func (m *PanicRecoverMiddleware) Process(app *Application, c *Context, next func() error) (err error) {
	defer func() {
		defer func() {
			if perr := recover(); perr != nil {
				app.logStackAndError(perr)
				err = fmt.Errorf("%v", perr)
			}
		}()
		if err != nil {
			app.Logger.Error(err)
			goto ERROR
		} else if perr := recover(); perr != nil {
			app.logStackAndError(perr)
			goto ERROR
		}
		return
	ERROR:
		c.Response.reset()
		if err = internalServerErrorController.GET(c); err != nil {
			app.logStackAndError(err)
		}
	}()
	return next()
}

// SessionMiddleware is a middleware to process a session.
type SessionMiddleware struct {
	// Name of cookie (key)
	Name string

	// Implementation of session store
	Store SessionStore

	// Expiration of session cookie, in seconds, from now. (not session expiration)
	// 0 is for persistent.
	CookieExpires time.Duration

	// Expiration of session data, in seconds, from now. (not cookie expiration)
	// 0 is for persistent.
	SessionExpires time.Duration
	HttpOnly       bool
}

func (m *SessionMiddleware) Process(app *Application, c *Context, next func() error) error {
	if err := m.before(app, c); err != nil {
		return err
	}
	if err := next(); err != nil {
		return err
	}
	return m.after(app, c)
}

// Validate validates configuration of the session.
func (m *SessionMiddleware) Validate() error {
	if m == nil {
		return fmt.Errorf("kocha: session: middleware is nil")
	}
	if m.Store == nil {
		return fmt.Errorf("kocha: session: because Store is nil, session cannot be used")
	}
	if m.Name == "" {
		return fmt.Errorf("kocha: session: Name must be specified")
	}
	return m.Store.Validate()
}

func (m *SessionMiddleware) before(app *Application, c *Context) (err error) {
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
	cookie, err := c.Request.Cookie(m.Name)
	if err != nil {
		return NewErrSession("new session")
	}
	sess, err := m.Store.Load(cookie.Value)
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

func (m *SessionMiddleware) after(app *Application, c *Context) (err error) {
	expires, _ := m.expiresFromDuration(m.SessionExpires)
	c.Session[SessionExpiresKey] = strconv.FormatInt(expires.Unix(), 10)
	cookie := m.newSessionCookie(app, c)
	cookie.Value, err = m.Store.Save(c.Session)
	if err != nil {
		return err
	}
	c.Response.SetCookie(cookie)
	return nil
}

func (m *SessionMiddleware) newSessionCookie(app *Application, c *Context) *http.Cookie {
	expires, maxAge := m.expiresFromDuration(m.CookieExpires)
	return &http.Cookie{
		Name:     m.Name,
		Value:    "",
		Path:     "/",
		Expires:  expires,
		MaxAge:   maxAge,
		Secure:   c.Request.IsSSL(),
		HttpOnly: m.HttpOnly,
	}
}

func (m *SessionMiddleware) expiresFromDuration(d time.Duration) (expires time.Time, maxAge int) {
	switch d {
	case -1:
		// persistent
		expires = util.Now().UTC().AddDate(20, 0, 0)
	case 0:
		expires = time.Time{}
	default:
		expires = util.Now().UTC().Add(d)
		maxAge = int(d.Seconds())
	}
	return expires, maxAge
}

// Flash messages processing middleware.
type FlashMiddleware struct{}

func (m *FlashMiddleware) Process(app *Application, c *Context, next func() error) error {
	if err := m.before(app, c); err != nil {
		return err
	}
	if err := next(); err != nil {
		return err
	}
	return m.after(app, c)
}

func (m *FlashMiddleware) before(app *Application, c *Context) error {
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

func (m *FlashMiddleware) after(app *Application, c *Context) error {
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

func (m *RequestLoggingMiddleware) Process(app *Application, c *Context, next func() error) error {
	defer func() {
		app.Logger.With(log.Fields{
			"method":   c.Request.Method,
			"uri":      c.Request.RequestURI,
			"protocol": c.Request.Proto,
			"status":   c.Response.StatusCode,
		}).Info()
	}()
	return next()
}

package kocha

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/naoina/kocha/util"
	"github.com/ugorji/go/codec"
)

// SessionConfig represents a configuration of session.
type SessionConfig struct {
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

func (config *SessionConfig) Validate() error {
	var sm *SessionMiddleware
	for _, m := range appConfig.Middlewares {
		if middleware, ok := m.(*SessionMiddleware); ok {
			sm = middleware
		}
	}
	if sm != nil {
		if config == nil {
			return fmt.Errorf("Because %T is nil, %T cannot be used", config, *sm)
		}
		if config.Store == nil {
			return fmt.Errorf("Because %T.Store is nil, %T cannot be used", *config, *sm)
		}
	}
	if config == nil {
		return nil
	}
	if config.Name == "" {
		return fmt.Errorf("%T.Name must be specify", *config)
	}
	if config.Store != nil {
		if err := config.Store.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// SessionStore is the interface that session store.
type SessionStore interface {
	Save(sess Session) (key string)
	Load(key string) (sess Session)

	// Validate calls in boot time.
	// Validate the session store specific values if you want. But highly recommended.
	Validate() error
}

// Session represents a session data store.
type Session map[string]string

const (
	SessionExpiresKey = "_kocha._sess._expires"
)

// Clear clear the all session data.
func (sess Session) Clear() {
	for k, _ := range sess {
		delete(sess, k)
	}
}

func newSessionCookie(c *Controller) *http.Cookie {
	expires, maxAge := expiresFromDuration(appConfig.Session.CookieExpires)
	return &http.Cookie{
		Name:     appConfig.Session.Name,
		Value:    "",
		Path:     "/",
		Expires:  expires,
		MaxAge:   maxAge,
		Secure:   c.Request.IsSSL(),
		HttpOnly: appConfig.Session.HttpOnly,
	}
}

func expiresFromDuration(d time.Duration) (expires time.Time, maxAge int) {
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

type ErrSession struct {
	msg string
}

func (e ErrSession) Error() string {
	return e.msg
}

func NewErrSession(msg string) error {
	return ErrSession{
		msg: msg,
	}
}

type ErrSessionExpected struct {
	msg string
}

func (e ErrSessionExpected) Error() string {
	return e.msg
}

func NewErrSessionExpected(msg string) error {
	return ErrSessionExpected{
		msg: msg,
	}
}

// Implementation of cookie store.
//
// This session store will be a session save to client-side cookie.
// Session cookie for save is encoded, encrypted and signed.
type SessionCookieStore struct {
	// key for the encryption.
	SecretKey string

	// Key for the cookie singing.
	SigningKey string
}

// Save saves and returns the key of session cookie.
// Actually, key is session cookie data itself.
func (store *SessionCookieStore) Save(sess Session) (key string) {
	var buf bytes.Buffer
	if err := codec.NewEncoder(&buf, &codec.MsgpackHandle{}).Encode(sess); err != nil {
		panic(err)
	}
	encrypted, err := store.encrypt(buf.Bytes())
	if err != nil {
		panic(err)
	}
	return store.encode(store.sign(encrypted))
}

// Load returns the session data that extract from cookie value.
// The key is stored session cookie value.
func (store *SessionCookieStore) Load(key string) (sess Session) {
	defer func() {
		if err := recover(); err != nil {
			if err, ok := err.(error); ok {
				panic(NewErrSession(err.Error()))
			}
			panic(err)
		}
	}()
	decoded, err := store.decode(key)
	if err != nil {
		panic(err)
	}
	unsigned, err := store.verify(decoded)
	if err != nil {
		panic(err)
	}
	decrypted, err := store.decrypt(unsigned)
	if err != nil {
		panic(err)
	}
	if err := codec.NewDecoderBytes(decrypted, &codec.MsgpackHandle{}).Decode(&sess); err != nil {
		panic(err)
	}
	return sess
}

// Validate validates SecretKey size.
func (store *SessionCookieStore) Validate() error {
	switch len(store.SecretKey) {
	case 16, 24, 32:
		return nil
	}
	return fmt.Errorf("%T.SecretKey size must be 16, 24 or 32, but %v", *store, len(store.SecretKey))
}

// encrypt returns encrypted data by AES-256-CBC.
func (store *SessionCookieStore) encrypt(buf []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(store.SecretKey))
	if err != nil {
		return nil, err
	}
	// padding for CBC
	rem := (aes.BlockSize - len(buf)%aes.BlockSize) % aes.BlockSize
	for i := 0; i < rem; i++ {
		buf = append(buf, byte(rem))
	}
	encrypted := make([]byte, aes.BlockSize+len(buf))
	iv := encrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted[aes.BlockSize:], buf)
	return encrypted, nil
}

// decrypt returns decrypted data from crypted data by AES-256-CBC.
func (store *SessionCookieStore) decrypt(buf []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(store.SecretKey))
	if err != nil {
		return nil, err
	}
	iv := buf[:aes.BlockSize]
	decrypted := buf[aes.BlockSize:]
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decrypted, decrypted)
	return decrypted, nil
}

// encode returns encoded string by Base64 with URLEncoding.
// However, encoded string will stripped the padding character of Base64.
func (store *SessionCookieStore) encode(src []byte) string {
	buf := make([]byte, base64.URLEncoding.EncodedLen(len(src)))
	base64.URLEncoding.Encode(buf, src)
	for {
		if buf[len(buf)-1] != '=' {
			break
		}
		buf = buf[:len(buf)-1]
	}
	return string(buf)
}

// decode returns decoded data from encoded data by Base64 with URLEncoding.
func (store *SessionCookieStore) decode(src string) ([]byte, error) {
	size := len(src)
	rem := (4 - size%4) % 4
	buf := make([]byte, size+rem)
	copy(buf, src)
	for i := 0; i < rem; i++ {
		buf[size+i] = '='
	}
	n, err := base64.URLEncoding.Decode(buf, buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// sign returns signed data.
func (store *SessionCookieStore) sign(src []byte) []byte {
	sign := store.hash(src)
	return append(sign, src...)
}

// verify verify signed data and returns unsigned data if valid.
func (store *SessionCookieStore) verify(src []byte) (unsigned []byte, err error) {
	if len(src) <= sha1.Size {
		return nil, errors.New("session cookie value too short")
	}
	sign := src[:sha1.Size]
	unsigned = src[sha1.Size:]
	if !hmac.Equal(store.hash(unsigned), sign) {
		return nil, errors.New("session cookie verification failed")
	}
	return unsigned, nil
}

// hash returns hashed data by HMAC-SHA1.
func (store *SessionCookieStore) hash(src []byte) []byte {
	hash := hmac.New(sha1.New, []byte(store.SigningKey))
	hash.Write(src)
	return hash.Sum(nil)
}

// Generate a random bytes.
func GenerateRandomKey(length int) []byte {
	result := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, result); err != nil {
		panic(err)
	}
	return result
}

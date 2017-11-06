package kocha

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/ugorji/go/codec"
)

// SessionStore is the interface that session store.
type SessionStore interface {
	Save(sess Session) (key string, err error)
	Load(key string) (sess Session, err error)
}

// Session represents a session data store.
type Session map[string]string

// Get gets a value associated with the given key.
// If there is the no value associated with the given key, Get returns "".
func (sess Session) Get(key string) string {
	return sess[key]
}

// Set sets the value associated with the key.
// If replaces the existing value associated with the key.
func (sess Session) Set(key, value string) {
	sess[key] = value
}

// Del deletes the value associated with the key.
func (sess Session) Del(key string) {
	delete(sess, key)
}

// Clear clear the all session data.
func (sess Session) Clear() {
	for k, _ := range sess {
		delete(sess, k)
	}
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

var codecHandler = &codec.MsgpackHandle{}

// Save saves and returns the key of session cookie.
// Actually, key is session cookie data itself.
func (store *SessionCookieStore) Save(sess Session) (key string, err error) {
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()
	if err := codec.NewEncoder(buf, codecHandler).Encode(sess); err != nil {
		return "", err
	}
	encrypted, err := store.encrypt(buf.Bytes())
	if err != nil {
		return "", err
	}
	return store.encode(store.sign(encrypted)), nil
}

// Load returns the session data that extract from cookie value.
// The key is stored session cookie value.
func (store *SessionCookieStore) Load(key string) (sess Session, err error) {
	decoded, err := store.decode(key)
	if err != nil {
		return nil, err
	}
	unsigned, err := store.verify(decoded)
	if err != nil {
		return nil, err
	}
	decrypted, err := store.decrypt(unsigned)
	if err != nil {
		return nil, err
	}
	if err := codec.NewDecoderBytes(decrypted, codecHandler).Decode(&sess); err != nil {
		return nil, err
	}
	return sess, nil
}

// Validate validates SecretKey size.
func (store *SessionCookieStore) Validate() error {
	b, err := base64.StdEncoding.DecodeString(store.SecretKey)
	if err != nil {
		return err
	}
	store.SecretKey = string(b)
	b, err = base64.StdEncoding.DecodeString(store.SigningKey)
	if err != nil {
		return err
	}
	store.SigningKey = string(b)
	switch len(store.SecretKey) {
	case 16, 24, 32:
		return nil
	}
	return fmt.Errorf("kocha: session: %T.SecretKey size must be 16, 24 or 32, but %v", *store, len(store.SecretKey))
}

// encrypt returns encrypted data by AES-256-CBC.
func (store *SessionCookieStore) encrypt(buf []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(store.SecretKey))
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, aead.NonceSize(), len(buf)+aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	encrypted := aead.Seal(nil, iv, buf, nil)
	return append(iv, encrypted...), nil
}

// decrypt returns decrypted data from crypted data by AES-256-CBC.
func (store *SessionCookieStore) decrypt(buf []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(store.SecretKey))
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	iv := buf[:aead.NonceSize()]
	decrypted := buf[aead.NonceSize():]
	if _, err := aead.Open(decrypted[:0], iv, decrypted, nil); err != nil {
		return nil, err
	}
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
	if len(src) <= sha512.Size256 {
		return nil, errors.New("kocha: session cookie value too short")
	}
	sign := src[:sha512.Size256]
	unsigned = src[sha512.Size256:]
	if !hmac.Equal(store.hash(unsigned), sign) {
		return nil, errors.New("kocha: session cookie verification failed")
	}
	return unsigned, nil
}

// hash returns hashed data by HMAC-SHA512/256.
func (store *SessionCookieStore) hash(src []byte) []byte {
	hash := hmac.New(sha512.New512_256, []byte(store.SigningKey))
	hash.Write(src)
	return hash.Sum(nil)
}

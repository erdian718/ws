package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"net/http"
	"sync"

	"github.com/ofunc/ws"
)

// Manager is the token manager.
type Manager struct {
	name   string
	path   string
	secure bool
	value  func() interface{}
	pool   *sync.Pool
}

// New creates a new token manager.
func New(name string, path string, secure bool, key []byte, value func() interface{}) *Manager {
	return &Manager{
		name:   name,
		path:   path,
		secure: secure,
		value:  value,
		pool: &sync.Pool{
			New: func() interface{} {
				return &token{
					buf:  make([]byte, 512),
					bbuf: make([]byte, 512),
					hs:   hmac.New(sha256.New, key),
				}
			},
		},
	}
}

// Create creates a new token.
func (a *Manager) Create(ctx *ws.Context, age int, value interface{}) error {
	t := a.pool.Get().(*token)
	defer a.pool.Put(t)
	if err := t.reset(age, value); err != nil {
		return err
	}
	http.SetCookie(ctx.ResponseWriter, &http.Cookie{
		Name:     a.name,
		Value:    t.String(),
		MaxAge:   age,
		Path:     a.path,
		Secure:   a.secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	return nil
}

// Delete deletes the token.
func (a *Manager) Delete(ctx *ws.Context) {
	http.SetCookie(ctx.ResponseWriter, &http.Cookie{
		Name:     a.name,
		MaxAge:   -1,
		Path:     a.path,
		Secure:   a.secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// Checker returns the token checker.
func (a *Manager) Checker(renew bool) func(*ws.Context) error {
	return func(ctx *ws.Context) error {
		if cookie, err := ctx.Request.Cookie(a.name); err == nil {
			t := a.pool.Get().(*token)
			defer a.pool.Put(t)
			value := a.value()
			if age, err := t.decode(cookie.Value, value, renew); err == nil {
				if renew {
					http.SetCookie(ctx.ResponseWriter, &http.Cookie{
						Name:     a.name,
						Value:    t.String(),
						MaxAge:   age,
						Path:     a.path,
						Secure:   a.secure,
						HttpOnly: true,
						SameSite: http.SameSiteStrictMode,
					})
				}
				ctx.Set(a.name, value)
				return ctx.Next()
			}
		}
		a.Delete(ctx)
		return ws.Status(http.StatusUnauthorized, "")
	}
}

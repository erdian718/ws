package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"hash"
	"net/http"
	"sync"
	"time"

	"github.com/ofunc/ws"
)

var (
	endian   = binary.LittleEndian
	encoding = base64.RawURLEncoding
)

// Manager is the token manager.
type Manager struct {
	name   string
	path   string
	secure bool
	pool   *sync.Pool
}

// New creates a new token manager.
func New(name string, path string, secure bool, key []byte) *Manager {
	return &Manager{
		name:   name,
		path:   path,
		secure: secure,
		pool: &sync.Pool{
			New: func() interface{} {
				return hmac.New(sha256.New, key)
			},
		},
	}
}

// New creates a new token.
func (a *Manager) New(ctx *ws.Context, age int32, value []byte) {
	http.SetCookie(ctx.ResponseWriter, &http.Cookie{
		Name:     a.name,
		Value:    a.encode(age, value),
		MaxAge:   int(age),
		Path:     a.path,
		Secure:   a.secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
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
			if ok, upt, age, value := a.decode(cookie.Value); ok {
				if renew && upt {
					a.New(ctx, age, value)
				}
				ctx.Set(a.name, value)
				return ctx.Next()
			}
		}
		a.Delete(ctx)
		return ws.Status(http.StatusUnauthorized, "")
	}
}

func (a *Manager) encode(age int32, value []byte) string {
	buf := make([]byte, 44+len(value))
	endian.PutUint64(buf[:8], uint64(time.Now().Unix()))
	endian.PutUint32(buf[8:], uint32(age))

	h := a.pool.Get().(hash.Hash)
	defer a.pool.Put(h)
	h.Reset()

	h.Write(buf[:12])
	h.Write(value)
	copy(buf[12:], h.Sum(nil))
	copy(buf[44:], value)
	return encoding.EncodeToString(buf)
}

func (a *Manager) decode(s string) (bool, bool, int32, []byte) {
	buf, err := encoding.DecodeString(s)
	if err != nil || len(buf) < 44 {
		return false, false, 0, nil
	}

	h := a.pool.Get().(hash.Hash)
	defer a.pool.Put(h)
	h.Reset()

	h.Write(buf[:12])
	h.Write(buf[44:])
	if !hmac.Equal(h.Sum(nil), buf[12:44]) {
		return false, false, 0, nil
	}

	stamp := int64(endian.Uint64(buf[:8]))
	age := int64(endian.Uint32(buf[8:12]))
	dur := time.Now().Unix() - stamp
	return dur <= age, dur > age/2, int32(age), buf[44:]
}

package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"hash"
	"net/http"
	"sync"
	"time"

	"github.com/ofunc/ws"
)

type itoken struct {
	stamp int64
	age   int32
}

// Manager is the token manager.
type Manager struct {
	name   string
	key    []byte
	path   string
	secure bool
	pool   sync.Pool
}

// Stamp gets the stamp.
func Stamp(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b[:8]))
}

// Value gets the value.
func Value(b []byte) []byte {
	return b[44:]
}

// New creates a new manager.
func New(name string, key []byte, path string, secure bool) *Manager {
	return &Manager{
		name:   name,
		key:    key,
		path:   path,
		secure: secure,
		pool: sync.Pool{
			New: func() interface{} {
				return hmac.New(sha256.New, key)
			},
		},
	}
}

// New creates a new token.
func (a *Manager) New(c *ws.Context, age int32, value []byte) {
	http.SetCookie(c.ResponseWriter, &http.Cookie{
		Name:     a.name,
		Value:    a.encode(age, value),
		MaxAge:   2 * int(age),
		Path:     a.path,
		Secure:   a.secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// Delete deletes the token.
func (a *Manager) Delete(c *ws.Context) {
	http.SetCookie(c.ResponseWriter, &http.Cookie{
		Name:     a.name,
		MaxAge:   -1,
		Path:     a.path,
		Secure:   a.secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// Check checks if the token is valid.
func (a *Manager) Check(c *ws.Context) error {
	if cookie, err := c.Request.Cookie(a.name); err == nil {
		if ok, age, buf := a.decode(cookie.Value); ok {
			if age > 0 {
				a.New(c, age, buf[44:])
			}
			c.Set(a.name, buf)
			return c.Next()
		}
	}
	a.Delete(c)
	return ws.Status(http.StatusUnauthorized, "")
}

func (a *Manager) encode(age int32, value []byte) string {
	buf := make([]byte, 12)
	binary.LittleEndian.PutUint64(buf[:8], uint64(time.Now().Unix()))
	binary.LittleEndian.PutUint32(buf[8:], uint32(age))

	h := a.pool.Get().(hash.Hash)
	defer a.pool.Put(h)
	h.Reset()

	h.Write(buf)
	h.Write(value)
	head := h.Sum(buf)
	return hex.EncodeToString(head) + hex.EncodeToString(value)
}

func (a *Manager) decode(s string) (bool, int32, []byte) {
	buf, err := hex.DecodeString(s)
	if err != nil {
		return false, 0, nil
	}
	if len(buf) < 44 {
		return false, 0, nil
	}

	h := a.pool.Get().(hash.Hash)
	defer a.pool.Put(h)
	h.Reset()

	h.Write(buf[:12])
	h.Write(buf[44:])
	if !hmac.Equal(h.Sum(nil), buf[12:44]) {
		return false, 0, nil
	}

	stamp := int64(binary.LittleEndian.Uint64(buf[:8]))
	age := int64(binary.LittleEndian.Uint32(buf[8:12]))
	dur := time.Now().Unix() - stamp
	if dur > 2*age {
		return false, 0, nil
	}
	if dur > age {
		return true, int32(age), buf
	}
	return true, 0, buf
}

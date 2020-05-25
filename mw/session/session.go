package session

import (
	"crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ofunc/ws"
)

// Session is the session entity.
type Session struct {
	Token    string
	Time     time.Time
	Duration time.Duration
}

// Manager is the session manager.
type Manager struct {
	sync.Mutex
	secure    bool
	path      string
	key       string
	idKey     string
	tokenKey  string
	clearHour int
	file      string
	store     map[string]*Session
}

// New creates a session manager.
func New() *Manager {
	mgr := &Manager{
		secure:   true,
		path:     "/",
		key:      "session",
		idKey:    "SESSION_ID",
		tokenKey: "SESSION_TOKEN",
		store:    make(map[string]*Session),
	}

	go func() {
		time.Sleep(1 * time.Hour)
		for {
			mgr.clear()
		}
	}()
	return mgr
}

// Secure sets the secure option.
func (a *Manager) Secure(o bool) *Manager {
	a.secure = o
	return a
}

// Key sets the key option.
func (a *Manager) Key(o string) *Manager {
	a.key = o
	return a
}

// IDKey sets the id key option.
func (a *Manager) IDKey(o string) *Manager {
	a.idKey = o
	return a
}

// TokenKey sets the token key option.
func (a *Manager) TokenKey(o string) *Manager {
	a.tokenKey = o
	return a
}

// ClearHour sets the clear hour option.
func (a *Manager) ClearHour(o int) *Manager {
	a.clearHour = o
	return a
}

// File sets the file option.
func (a *Manager) File(o string) *Manager {
	a.file = o
	if a.file != "" {
		f, err := os.Open(a.file)
		if err != nil {
			return a
		}
		defer f.Close()

		store := make(map[string]*Session)
		if err := gob.NewDecoder(f).Decode(&store); err == nil {
			a.store = store
		}
	}
	return a
}

// New creates a new session.
func (a *Manager) New(c *ws.Context, id string, duration time.Duration) error {
	if id == "" {
		return errors.New("session: invalid session id")
	}

	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		return err
	}
	stoken := hex.EncodeToString(token)

	http.SetCookie(c.ResponseWriter, &http.Cookie{
		Name:     a.idKey,
		Value:    id,
		Path:     a.path,
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(c.ResponseWriter, &http.Cookie{
		Name:     a.tokenKey,
		Value:    stoken,
		Path:     a.path,
		SameSite: http.SameSiteStrictMode,
		Secure:   a.secure,
		HttpOnly: true,
	})

	s := &Session{
		Token:    stoken,
		Duration: duration,
		Time:     time.Now().Add(duration),
	}
	a.Lock()
	a.store[id] = s
	a.Unlock()
	c.Set(a.key, id)
	return nil
}

// Delete deletes the session.
func (a *Manager) Delete(c *ws.Context, id string) {
	a.deleteCookie(c)
	a.Lock()
	delete(a.store, id)
	a.Unlock()
}

// Auth auths the session.
func (a *Manager) Auth(c *ws.Context) error {
	cid, err := c.Request.Cookie(a.idKey)
	if err != nil {
		a.deleteCookie(c)
		return c.Status(http.StatusUnauthorized)
	}
	id := cid.Value

	ctoken, err := c.Request.Cookie(a.tokenKey)
	if err != nil {
		a.deleteCookie(c)
		return c.Status(http.StatusUnauthorized)
	}
	token := ctoken.Value

	now := time.Now()
	a.Lock()
	s, ok := a.store[id]
	if ok {
		if ok = now.Before(s.Time); ok {
			if ok = s.Token == token; ok {
				s.Time = now.Add(s.Duration)
			}
		} else {
			delete(a.store, id)
		}
	}
	a.Unlock()

	if ok {
		c.Set(a.key, id)
		return c.Next()
	}
	a.deleteCookie(c)
	return c.Status(http.StatusUnauthorized)
}

func (a *Manager) deleteCookie(c *ws.Context) {
	http.SetCookie(c.ResponseWriter, &http.Cookie{
		Name:     a.idKey,
		Value:    "",
		Path:     a.path,
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(c.ResponseWriter, &http.Cookie{
		Name:     a.tokenKey,
		Value:    "",
		Path:     a.path,
		SameSite: http.SameSiteStrictMode,
		Secure:   a.secure,
		HttpOnly: true,
	})
}

func (a *Manager) clear() {
	defer func() {
		if e := recover(); e != nil {
			log.Println("session:", e)
		}
	}()

	now := time.Now()
	time.Sleep(time.Date(now.Year(), now.Month(), now.Day()+1, a.clearHour, 0, 0, 0, time.Local).Sub(now))
	now = time.Now()

	a.Lock()
	defer a.Unlock()

	for id, s := range a.store {
		if s.Time.Before(now) {
			delete(a.store, id)
		}
	}

	if a.file != "" {
		f, err := os.Create(a.file)
		if err != nil {
			return
		}
		defer f.Close()
		gob.NewEncoder(f).Encode(a.store)
	}
}

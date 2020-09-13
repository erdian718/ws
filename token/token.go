package token

import (
	"crypto/hmac"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash"
	"time"
)

// Errors.
var (
	ErrInvalidToken = errors.New("ws/token: invalid token")
	ErrExpiredToken = errors.New("ws/token: expired token")
)

var (
	endian   = binary.LittleEndian
	encoding = base64.RawURLEncoding
)

type token struct {
	buf  []byte
	bbuf []byte
	hs   hash.Hash
}

// String returns the token as string.
func (a *token) String() string {
	n := encoding.EncodedLen(len(a.buf))
	if len(a.bbuf) < n {
		a.bbuf = make([]byte, n)
	}
	encoding.Encode(a.bbuf, a.buf)
	return string(a.bbuf[:n])
}

func (a *token) sign() {
	a.hs.Reset()
	a.hs.Write(a.buf[:12])
	a.hs.Write(a.buf[44:])
	copy(a.buf[12:], a.hs.Sum(nil))
}

func (a *token) reset(age int, v interface{}) error {
	endian.PutUint64(a.buf[:8], uint64(time.Now().Unix()))
	endian.PutUint32(a.buf[8:], uint32(age))

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	n := 44 + len(data)
	if cap(a.buf) < n {
		buf := make([]byte, n)
		copy(buf, a.buf[:12])
		a.buf = buf
	} else {
		a.buf = a.buf[:n]
	}
	copy(a.buf[44:], data)
	a.sign()
	return nil
}

func (a *token) decode(s string, value interface{}, renew bool) (int, error) {
	n := encoding.DecodedLen(len(s))
	if n < 44 {
		return 0, ErrInvalidToken
	}
	if cap(a.buf) < n {
		a.buf = make([]byte, n)
	} else {
		a.buf = a.buf[:n]
	}
	if _, err := encoding.Decode(a.buf, []byte(s)); err != nil {
		return 0, err
	}

	a.hs.Reset()
	a.hs.Write(a.buf[:12])
	a.hs.Write(a.buf[44:])
	if !hmac.Equal(a.hs.Sum(nil), a.buf[12:44]) {
		return 0, ErrInvalidToken
	}

	now := time.Now().Unix()
	stamp := int64(endian.Uint64(a.buf[:8]))
	age := int64(endian.Uint32(a.buf[8:12]))
	dur := now - stamp
	if dur > age {
		return int(age), ErrExpiredToken
	}
	if renew && dur > age/2 {
		endian.PutUint64(a.buf[:8], uint64(now))
		a.sign()
	}
	return int(age), json.Unmarshal(a.buf[44:], value)
}

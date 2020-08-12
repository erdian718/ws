package token

import (
	"crypto/hmac"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"hash"
	"io"
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
	rindex int
	windex int
	buf    []byte
	bbuf   []byte
	dec    *gob.Decoder
	enc    *gob.Encoder
	hs     hash.Hash
}

// Write implements the io.Writer interface.
func (a *token) Write(p []byte) (n int, err error) {
	n = len(p)
	c := a.windex + n
	if c > len(a.buf) {
		buf := make([]byte, c+a.windex)
		copy(buf, a.buf[:a.windex])
		a.buf = buf
	}
	copy(a.buf[a.windex:], p)
	a.windex = c
	return
}

// Read implements the io.Readr interface.
func (a *token) Read(p []byte) (n int, err error) {
	n = copy(p, a.buf[a.rindex:a.windex])
	a.rindex += n
	if n < len(p) {
		err = io.EOF
	}
	return
}

// String returns the token as string.
func (a *token) String() string {
	n := encoding.EncodedLen(a.windex)
	if len(a.bbuf) < n {
		a.bbuf = make([]byte, n)
	}
	encoding.Encode(a.bbuf, a.buf[:a.windex])
	return string(a.bbuf[:n])
}

func (a *token) sign() {
	a.hs.Reset()
	a.hs.Write(a.buf[:12])
	a.hs.Write(a.buf[44:a.windex])
	copy(a.buf[12:], a.hs.Sum(nil))
}

func (a *token) reset(age int, v interface{}) error {
	endian.PutUint64(a.buf[:8], uint64(time.Now().Unix()))
	endian.PutUint32(a.buf[8:], uint32(age))
	a.windex = 44
	err := a.enc.Encode(v)
	a.sign()
	return err
}

func (a *token) decode(s string, value interface{}, renew bool) (int, error) {
	a.windex = encoding.DecodedLen(len(s))
	if len(a.buf) < a.windex {
		a.buf = make([]byte, a.windex)
	}
	if _, err := encoding.Decode(a.buf, []byte(s)); err != nil {
		return 0, err
	}

	a.hs.Reset()
	a.hs.Write(a.buf[:12])
	a.hs.Write(a.buf[44:a.windex])
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
	a.rindex = 44
	return int(age), a.dec.Decode(value)
}

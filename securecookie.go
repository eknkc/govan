package govan

import (
	"github.com/gorilla/securecookie"
)

type SecureCookie interface {
	Encode(name string, value interface{}) (string, error)
	Decode(name string, value string, target interface{}) error
}

type secureCookie struct {
	sc *securecookie.SecureCookie
}

func (s *secureCookie) Encode(name string, value interface{}) (string, error) {
	return s.sc.Encode(name, value)
}

func (s *secureCookie) Decode(name string, value string, target interface{}) error {
	return s.sc.Decode(name, value, target)
}

func NewSecureCookie(hashKey, blockKey string) SecureCookie {
	return &secureCookie{
		sc: securecookie.New([]byte(hashKey), []byte(blockKey)),
	}
}

package httper

import (
	"net/http"
	"time"
)

// Cookier defines a cookie provider requirements.
type Cookier interface {
	Get(name string) string
	GetCookie(name string) *http.Cookie
	Set(name string, value string, t ...time.Time) *http.Cookie
	SetCookie(cookie *http.Cookie)
}

// CookieProvider is a Cookier factory.
type CookieProvider interface {
	Make(w http.ResponseWriter, r *http.Request) Cookier
}

// CookieHelperProvider instancatiates CookieHelpers.
type CookieHelperProvider struct{}

// Make returns a CookieHelper
func (c CookieHelperProvider) Make(w http.ResponseWriter, r *http.Request) Cookier {
	return &CookieHelper{w, r}
}

// CookieHelper helps to deal with cookies.
type CookieHelper struct {
	w http.ResponseWriter
	r *http.Request
}

// Get a cookie value
func (c CookieHelper) Get(name string) string {
	cookie, _ := c.r.Cookie(name)
	if cookie == nil {
		return ""
	}
	return cookie.Value
}

// GetCookie returns a plain cookie.
func (c CookieHelper) GetCookie(name string) *http.Cookie {
	cookie, _ := c.r.Cookie(name)
	return cookie
}

// Set a cookie value
func (c CookieHelper) Set(name string, value string, t ...time.Time) *http.Cookie {
	expiration := time.Now().Add(365 * 24 * time.Hour)
	if len(t) > 0 {
		expiration = t[0]
	}
	cookie := &http.Cookie{Name: name, Value: value, Expires: expiration}
	http.SetCookie(c.w, cookie)
	return cookie
}

// SetCookie sets a plain cookie
func (c CookieHelper) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.w, cookie)
}

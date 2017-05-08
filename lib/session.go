package httper

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// Sessionner defines a session provider requirements.
type Sessionner interface {
	Get(name string) string
	GetAny(name string) interface{}
	Set(name string, value string)
	SetAny(name string, value interface{})
}

// SessionProvider is a Sessionner factory.
type SessionProvider interface {
	Make(w http.ResponseWriter, r *http.Request) Sessionner
}

// VoidSessionProvider instancatiates SessionProvider.
type VoidSessionProvider struct {
}

// Make returns a SessionProvider
func (c VoidSessionProvider) Make(w http.ResponseWriter, r *http.Request) Sessionner {
	return &VoidSession{w, r}
}

// VoidSession helps to deal with cookies.
type VoidSession struct {
	w http.ResponseWriter
	r *http.Request
}

// Get a string
func (c VoidSession) Get(name string) string {
	return ""
}

// GetAny returns any kind of value.
func (c VoidSession) GetAny(name string) interface{} {
	return nil
}

// SetAny kind of value.
func (c VoidSession) SetAny(name string, value interface{}) {}

// Set a string.
func (c VoidSession) Set(name, value string) {}

// GorillaSessionProvider instancatiates SessionProvider.
type GorillaSessionProvider struct {
	Name  string
	store sessions.Store
}

// Make returns a SessionProvider
func (c GorillaSessionProvider) Make(w http.ResponseWriter, r *http.Request) Sessionner {
	s, _ := c.store.Get(r, c.Name)
	//doc:
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	return &GorillaSession{w, r, s}
}

// GorillaSession helps to deal with cookies.
type GorillaSession struct {
	w       http.ResponseWriter
	r       *http.Request
	session *sessions.Session
}

// Get a string
func (c GorillaSession) Get(name string) string {
	return c.GetAny(name).(string)
}

// GetAny returns any kind of value.
func (c GorillaSession) GetAny(name string) interface{} {
	if val, ok := c.session.Values[name]; ok {
		return val
	}
	return nil
}

// SetAny kind of value.
func (c GorillaSession) SetAny(name string, value interface{}) {
	c.session.Values[name] = value
}

// Set a string.
func (c GorillaSession) Set(name, value string) {
	c.SetAny(name, value)
}

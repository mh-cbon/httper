package httper

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// Dataer defines a data provider requirements.
type Dataer interface {
	Get(prefix, name string) string
	GetAny(prefix, name string) interface{}
}

// DataerProvider is a Dataer factory.
type DataerProvider interface {
	Make(w http.ResponseWriter, r *http.Request) Dataer
	MakeEmpty() Dataer
}

//DataProvider isa Dataer able to tell about more a prefix
type DataProvider interface {
	Dataer
	IsAbout(string) bool
	GetName() string
}

//DataProviderFacade multiple providers
type DataProviderFacade struct {
	Providers []DataProvider
}

// NewDataProviderFacade constructs a data provider facade.
func NewDataProviderFacade(providers ...DataProvider) *DataProviderFacade {
	return &DataProviderFacade{
		Providers: providers,
	}
}

// Get a string
func (c *DataProviderFacade) Get(prefix, name string) string {
	return c.GetAny(prefix, name).(string)
}

// GetAny kind of value
func (c *DataProviderFacade) GetAny(prefix, name string) interface{} {
	var ret interface{}
	for _, p := range c.Providers {
		if p.IsAbout(prefix) {
			ret = p.Get(prefix, name)
			if ret != nil {
				break
			}
		}
	}
	return ret
}

// StdHTTPDataProvider returns a data provider for a standard http handlingr.
type StdHTTPDataProvider struct{}

// Make returns a DataHelper
func (c StdHTTPDataProvider) Make(w http.ResponseWriter, r *http.Request) Dataer {
	query := r.URL.Query()
	return NewDataProviderFacade(
		&GetHTTPDataProvider{w, r, query},
		&PostHTTPDataProvider{w, r},
		&CookieHTTPDataProvider{w, r},
		&ReqHTTPDataProvider{w, r, query},
	)
}

// MakeEmpty returns a DataHelper
func (c StdHTTPDataProvider) MakeEmpty() Dataer {
	return NewDataProviderFacade(
		&GetHTTPDataProvider{},
		&PostHTTPDataProvider{},
		&CookieHTTPDataProvider{},
		&ReqHTTPDataProvider{},
	)
}

// GorillaHTTPDataProvider returns a data provider for a standard http handlingr.
type GorillaHTTPDataProvider struct {
	StdHTTPDataProvider
	store sessions.Store
	Name  string
}

// Make returns a DataHelper
func (c GorillaHTTPDataProvider) Make(w http.ResponseWriter, r *http.Request) Dataer {
	ret := c.StdHTTPDataProvider.Make(w, r).(*DataProviderFacade)

	query := r.URL.Query()
	vars := mux.Vars(r)
	ret.Providers = append(ret.Providers,
		&URLHTTPDataProvider{w, r, query, vars},
		&RouteHTTPDataProvider{w, r, vars},
	)

	if c.store != nil {
		// from the doc:
		// Get a session. We're ignoring the error resulted from decoding an
		// existing session: Get() always returns a session, even if empty.
		s, _ := c.store.Get(r, c.Name)
		ret.Providers = append(
			ret.Providers,
			&GorillaSessionHTTPDataProvider{w, r, s},
		)
	}
	return ret
}

// MakeEmpty returns a DataHelper
func (c GorillaHTTPDataProvider) MakeEmpty() Dataer {
	ret := c.StdHTTPDataProvider.MakeEmpty().(*DataProviderFacade)
	ret.Providers = append(ret.Providers,
		&URLHTTPDataProvider{},
		&RouteHTTPDataProvider{},
		&GorillaSessionHTTPDataProvider{},
	)
	return ret
}

// GetHTTPDataProvider helps to deal with GET.
type GetHTTPDataProvider struct {
	w     http.ResponseWriter
	r     *http.Request
	query url.Values
}

// IsAbout returns true when prefix is get
func (c GetHTTPDataProvider) IsAbout(prefix string) bool {
	return prefix == c.GetName()
}

// GetName returns get
func (c GetHTTPDataProvider) GetName() string {
	return "get"
}

// Get a string
func (c GetHTTPDataProvider) Get(prefix, name string) string {
	if _, ok := c.query[name]; ok {
		return c.query.Get(name)
	}
	return ""
}

// GetAny kind of value
func (c GetHTTPDataProvider) GetAny(prefix, name string) interface{} {
	return c.Get(prefix, name)
}

// CookieHTTPDataProvider helps to deal with GET.
type CookieHTTPDataProvider struct {
	w http.ResponseWriter
	r *http.Request
}

// IsAbout returns true when prefix is get
func (c CookieHTTPDataProvider) IsAbout(prefix string) bool {
	return prefix == c.GetName()
}

// GetName returns cookie
func (c CookieHTTPDataProvider) GetName() string {
	return "cookie"
}

// Get a string
func (c CookieHTTPDataProvider) Get(prefix, name string) string {
	if cookie, _ := c.r.Cookie(name); cookie != nil {
		return cookie.Value
	}
	return ""
}

// GetAny kind of value
func (c CookieHTTPDataProvider) GetAny(prefix, name string) interface{} {
	return c.Get(prefix, name)
}

// ReqHTTPDataProvider helps to deal with Req.
type ReqHTTPDataProvider struct {
	w     http.ResponseWriter
	r     *http.Request
	query url.Values
}

// IsAbout returns true when prefix is get
func (c ReqHTTPDataProvider) IsAbout(prefix string) bool {
	return prefix == c.GetName()
}

// GetName returns req
func (c ReqHTTPDataProvider) GetName() string {
	return "req"
}

// Get a string
func (c ReqHTTPDataProvider) Get(prefix, name string) string {
	if _, ok := c.query[name]; ok {
		return c.query.Get(name)
	}
	if err := c.r.ParseForm(); err == nil {
		if _, ok := c.r.Form[name]; ok {
			return c.r.FormValue(name)
		}
	}
	return ""
}

// GetAny kind of value
func (c ReqHTTPDataProvider) GetAny(prefix, name string) interface{} {
	return c.Get(prefix, name)
}

// PostHTTPDataProvider helps to deal with GET.
type PostHTTPDataProvider struct {
	w http.ResponseWriter
	r *http.Request
}

// IsAbout returns true when prefix is get
func (c PostHTTPDataProvider) IsAbout(prefix string) bool {
	return prefix == c.GetName()
}

// GetName returns post
func (c PostHTTPDataProvider) GetName() string {
	return "post"
}

// Get a string
func (c PostHTTPDataProvider) Get(prefix, name string) string {
	if err := c.r.ParseForm(); err == nil {
		if _, ok := c.r.Form[name]; ok {
			return c.r.FormValue(name)
		}
	}
	return ""
}

// GetAny kind of value
func (c PostHTTPDataProvider) GetAny(prefix, name string) interface{} {
	return c.Get(prefix, name)
}

// URLHTTPDataProvider helps to deal with URL.
type URLHTTPDataProvider struct {
	w     http.ResponseWriter
	r     *http.Request
	query url.Values
	vars  map[string]string
}

// IsAbout returns true when prefix is get
func (c URLHTTPDataProvider) IsAbout(prefix string) bool {
	return prefix == c.GetName()
}

// GetName returns url
func (c URLHTTPDataProvider) GetName() string {
	return "url"
}

// Get a string
func (c URLHTTPDataProvider) Get(prefix, name string) string {
	if _, ok := c.query[name]; ok {
		return c.query.Get(name)
	}
	if val, ok := c.vars[name]; ok {
		return val
	}
	return ""
}

// GetAny kind of value
func (c URLHTTPDataProvider) GetAny(prefix, name string) interface{} {
	return c.Get(prefix, name)
}

// RouteHTTPDataProvider helps to deal with URL.
type RouteHTTPDataProvider struct {
	w    http.ResponseWriter
	r    *http.Request
	vars map[string]string
}

// IsAbout returns true when prefix is get
func (c RouteHTTPDataProvider) IsAbout(prefix string) bool {
	return prefix == c.GetName()
}

// GetName returns route
func (c RouteHTTPDataProvider) GetName() string {
	return "route"
}

// Get a string
func (c RouteHTTPDataProvider) Get(prefix, name string) string {
	if val, ok := c.vars[name]; ok {
		return val
	}
	return ""
}

// GetAny kind of value
func (c RouteHTTPDataProvider) GetAny(prefix, name string) interface{} {
	return c.Get(prefix, name)
}

// GorillaSessionHTTPDataProvider helps to deal with URL.
type GorillaSessionHTTPDataProvider struct {
	w       http.ResponseWriter
	r       *http.Request
	session *sessions.Session
}

// IsAbout returns true when prefix is get
func (c GorillaSessionHTTPDataProvider) IsAbout(prefix string) bool {
	return prefix == c.GetName()
}

// GetName returns session
func (c GorillaSessionHTTPDataProvider) GetName() string {
	return "session"
}

// Get a string
func (c GorillaSessionHTTPDataProvider) Get(prefix, name string) string {
	return c.GetAny(prefix, name).(string)
}

// GetAny kind of value
func (c GorillaSessionHTTPDataProvider) GetAny(prefix, name string) interface{} {
	if val, ok := c.session.Values[name]; ok {
		return val
	}
	return nil
}

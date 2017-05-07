package httper

import (
	"net/http"
)

// Dataer defines a data provider requirements.
type Dataer interface {
	Get(prefix, name string) string
}

// DataerProvider is a Dataer factory.
type DataerProvider interface {
	Make(w http.ResponseWriter, r *http.Request) Dataer
}

// DataHelperProvider instancatiates DataHelper.
type DataHelperProvider struct{}

// Make returns a DataHelper
func (c DataHelperProvider) Make(w http.ResponseWriter, r *http.Request) Dataer {
	return &DataHelper{w, r}
}

// DataHelper helps to deal with cookies.
type DataHelper struct {
	w http.ResponseWriter
	r *http.Request
}

// Get a value
func (c DataHelper) Get(prefix, name string) string {
	if prefix == "get" || prefix == "url" || prefix == "req" {
		reqValues := c.r.URL.Query()
		if _, ok := reqValues[name]; ok {
			return reqValues.Get(name)
		}
	}
	if prefix == "post" || prefix == "req" {
		if err := c.r.ParseForm(); err == nil {
			if _, ok := c.r.Form[name]; ok {
				return c.r.FormValue(name)
			}
		}
	}
	if prefix == "cookie" {
		if cookie, _ := c.r.Cookie(name); cookie != nil {
			return cookie.Value
		}
	}
	return ""
}

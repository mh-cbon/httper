package httper

import (
	"io"
	"net/http"
)

// Finalizer finalizes an htpp response.
type Finalizer interface {
	HandleError(err error, w io.Writer, r *http.Request) bool
	HandleSuccess(w io.Writer, r io.Reader) error
}

// DefaultFinalizer for an http response.
type DefaultFinalizer struct {
}

// HandleError prints http 500.
func (f DefaultFinalizer) HandleError(err error, w io.Writer, r *http.Request) bool {
	if x, ok := w.(http.ResponseWriter); ok {
		http.Error(x, err.Error(), http.StatusInternalServerError)
	}
	return true
}

// HandleSuccess prints http 200 and prints r.
func (f DefaultFinalizer) HandleSuccess(w io.Writer, r io.Reader) error {
	if x, ok := w.(http.ResponseWriter); ok {
		x.WriteHeader(http.StatusOK)
	}
	return nil
}

// HTTPFinalizer finalizes an HTTP response.
type HTTPFinalizer struct {
	DefaultFinalizer
}

// HandleSuccess prints http 200 and prints r.
func (f HTTPFinalizer) HandleSuccess(w io.Writer, r io.Reader) error {
	f.DefaultFinalizer.HandleSuccess(w, r)
	_, err := io.Copy(w, r)
	return err
}

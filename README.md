# httper

[![travis Status](https://travis-ci.org/mh-cbon/httper.svg?branch=master)](https://travis-ci.org/mh-cbon/httper) [![Appveyor Status](https://ci.appveyor.com/api/projects/status/github/mh-cbon/httper?branch=master&svg=true)](https://ci.appveyor.com/projects/mh-cbon/httper) [![Go Report Card](https://goreportcard.com/badge/github.com/mh-cbon/httper)](https://goreportcard.com/report/github.com/mh-cbon/httper) [![GoDoc](https://godoc.org/github.com/mh-cbon/httper?status.svg)](http://godoc.org/github.com/mh-cbon/httper) [![MIT License](http://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Package httper is a cli tool to implement http interface of a type.


s/Choose your gun!/[Aux armes!](https://www.youtube.com/watch?v=hD-wD_AMRYc&t=7)/

# TOC
- [Install](#install)
  - [glide](#glide)
- [Usage](#usage)
  - [$ httper -help](#-httper--help)
- [Cli examples](#cli-examples)
- [API example](#api-example)
  - [> demo/main.go](#-demomaingo)
  - [> demo/controllerhttpgen.go](#-democontrollerhttpgengo)
- [Recipes](#recipes)
  - [Release the project](#release-the-project)
- [History](#history)

# Install

#### glide
```sh
mkdir -p $GOPATH/src/github.com/mh-cbon/httper
cd $GOPATH/src/github.com/mh-cbon/httper
git clone https://github.com/mh-cbon/httper.git .
glide install
go install
```


# Usage

#### $ httper -help
```sh
httper 0.0.0

Usage

	httper [-p name] [-mode name] [...types]

  types:  A list of types such as src:dst.
          A type is defined by its package path and its type name,
          [pkgpath/]name
          If the Package path is empty, it is set to the package name being generated.
          Name can be a valid type identifier such as TypeName, *TypeName, []TypeName 
  -p:     The name of the package output.
  -mode:  The mode of generation to apply: std|gorilla (defaults to std).
```

## Cli examples

```sh
# Create a httped version of JSONTomates to HTTPTomates
httper *JSONTomates:HTTPTomates
# Create a jsoned version of JSONTomates to HTTPTomates to stdout
httper -p main - JSONTomates:HTTPTomates
```

# API example

Following example demonstates a program using it to generate an `httped` version of a type.

#### > demo/main.go
```go
package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"

	httper "github.com/mh-cbon/httper/lib"
)

//go:generate lister *Tomate:TomatesGen
//go:generate channeler TomatesGen:TomatesSyncGen

//go:generate jsoner -mode gorilla *Controller:ControllerJSONGen
//go:generate httper -mode gorilla *ControllerJSONGen:ControllerHTTPGen

func main() {

	backend := NewTomatesSyncGen()
	backend.Push(&Tomate{Name: "Red"})

	jsoner := NewControllerJSONGen(NewController(backend), nil)
	httper := NewControllerHTTPGen(jsoner, nil)
	// public views
	http.HandleFunc("/", httper.GetByID)

	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	time.Sleep(1 * time.Millisecond)

	req, err := http.Get("http://localhost:8080/?id=0")
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()
	io.Copy(os.Stdout, req.Body)

}

// Tomate is about red vegetables to make famous italian food.
type Tomate struct {
	ID   int
	Name string
}

// GetID return the ID of the Tomate.
func (t *Tomate) GetID() int {
	return t.ID
}

// Controller of some resources.
type Controller struct {
	backend *TomatesSyncGen
}

// NewController ...
func NewController(backend *TomatesSyncGen) *Controller {
	return &Controller{
		backend: backend,
	}
}

// GetByID ...
func (t *Controller) GetByID(urlID int) *Tomate {
	return t.backend.Filter(FilterTomatesGen.ByID(urlID)).First()
}

// UpdateByID ...
func (t *Controller) UpdateByID(urlID int, reqBody *Tomate) *Tomate {
	var ret *Tomate
	t.backend.Filter(func(v *Tomate) bool {
		if v.ID == urlID {
			v.Name = reqBody.Name
			ret = v
		}
		return true
	})
	return ret
}

// DeleteByID ...
func (t *Controller) DeleteByID(REQid int) bool {
	return t.backend.Remove(&Tomate{ID: REQid})
}

// TestVars1 ...
func (t *Controller) TestVars1(w http.ResponseWriter, r *http.Request) {
}

// TestCookier ...
func (t *Controller) TestCookier(c httper.Cookier) {
}

// TestSessionner ...
func (t *Controller) TestSessionner(s httper.Sessionner) {
}

// TestRPCer ...
func (t *Controller) TestRPCer(id int) bool {
	return false
}
```

Following code is the generated implementation of an `httped` typed slice of `Tomate`.

#### > demo/controllerhttpgen.go
```go
package main

// file generated by
// github.com/mh-cbon/httper
// do not edit

import (
	httper "github.com/mh-cbon/httper/lib"
	"io"
	"net/http"
	"strconv"
)

var xxStrconvAtoi = strconv.Atoi
var xxIoCopy = io.Copy
var xxHTTPOk = http.StatusOK

// ControllerHTTPGen is an httper of *ControllerJSONGen.
// ControllerJSONGen is jsoner of *Controller.
// Controller of some resources.
type ControllerHTTPGen struct {
	embed     *ControllerJSONGen
	cookier   httper.CookieProvider
	dataer    httper.DataerProvider
	sessioner httper.SessionProvider
	finalizer httper.Finalizer
}

// NewControllerHTTPGen constructs an httper of *ControllerJSONGen
func NewControllerHTTPGen(embed *ControllerJSONGen, finalizer httper.Finalizer) *ControllerHTTPGen {
	if finalizer == nil {
		finalizer = &httper.HTTPFinalizer{}
	}
	ret := &ControllerHTTPGen{
		embed:     embed,
		cookier:   &httper.CookieHelperProvider{},
		dataer:    &httper.GorillaHTTPDataProvider{},
		sessioner: &httper.GorillaSessionProvider{},
		finalizer: finalizer,
	}
	return ret
}

// GetByID invoke *ControllerJSONGen.GetByID using the request body as a json payload.
// GetByID Decodes reqBody as json to invoke *Controller.GetByID.
// Other parameters are passed straight
// GetByID ...
func (t *ControllerHTTPGen) GetByID(w http.ResponseWriter, r *http.Request) {
	var urlID int
	tempurlID, err := strconv.Atoi(t.dataer.Make(w, r).Get("url", "id"))
	if err != nil && t.finalizer.HandleError(err, w, r) {
		return
	}
	urlID = tempurlID

	res, err := t.embed.GetByID(urlID)
	if err != nil && t.finalizer.HandleError(err, w, r) {
		return
	}

	t.finalizer.HandleSuccess(w, res)

}

// UpdateByID invoke *ControllerJSONGen.UpdateByID using the request body as a json payload.
// UpdateByID Decodes reqBody as json to invoke *Controller.UpdateByID.
// Other parameters are passed straight
// UpdateByID ...
func (t *ControllerHTTPGen) UpdateByID(w http.ResponseWriter, r *http.Request) {
	var urlID int
	tempurlID, err := strconv.Atoi(t.dataer.Make(w, r).Get("url", "id"))
	if err != nil && t.finalizer.HandleError(err, w, r) {
		return
	}
	urlID = tempurlID
	reqBody := r.Body

	res, err := t.embed.UpdateByID(urlID, reqBody)
	if err != nil && t.finalizer.HandleError(err, w, r) {
		return
	}

	t.finalizer.HandleSuccess(w, res)

}

// DeleteByID invoke *ControllerJSONGen.DeleteByID using the request body as a json payload.
// DeleteByID Decodes reqBody as json to invoke *Controller.DeleteByID.
// Other parameters are passed straight
// DeleteByID ...
func (t *ControllerHTTPGen) DeleteByID(w http.ResponseWriter, r *http.Request) {
	var REQid int
	tempREQid, err := strconv.Atoi(t.dataer.Make(w, r).Get("req", "id"))
	if err != nil && t.finalizer.HandleError(err, w, r) {
		return
	}
	REQid = tempREQid

	res, err := t.embed.DeleteByID(REQid)
	if err != nil && t.finalizer.HandleError(err, w, r) {
		return
	}

	t.finalizer.HandleSuccess(w, res)

}

// TestVars1 invoke *ControllerJSONGen.TestVars1 using the request body as a json payload.
// TestVars1 Decodes reqBody as json to invoke *Controller.TestVars1.
// Other parameters are passed straight
// TestVars1 ...
func (t *ControllerHTTPGen) TestVars1(w http.ResponseWriter, r *http.Request) {

	res, err := t.embed.TestVars1(w, r)
	if err != nil && t.finalizer.HandleError(err, w, r) {
		return
	}

	t.finalizer.HandleSuccess(w, res)

}

// TestCookier invoke *ControllerJSONGen.TestCookier using the request body as a json payload.
// TestCookier Decodes reqBody as json to invoke *Controller.TestCookier.
// Other parameters are passed straight
// TestCookier ...
func (t *ControllerHTTPGen) TestCookier(w http.ResponseWriter, r *http.Request) {
	var c httper.Cookier
	c = t.cookier.Make(w, r)

	res, err := t.embed.TestCookier(c)
	if err != nil && t.finalizer.HandleError(err, w, r) {
		return
	}

	t.finalizer.HandleSuccess(w, res)

}

// TestSessionner invoke *ControllerJSONGen.TestSessionner using the request body as a json payload.
// TestSessionner Decodes reqBody as json to invoke *Controller.TestSessionner.
// Other parameters are passed straight
// TestSessionner ...
func (t *ControllerHTTPGen) TestSessionner(w http.ResponseWriter, r *http.Request) {
	var s httper.Sessionner
	s = t.sessioner.Make(w, r)

	res, err := t.embed.TestSessionner(s)
	if err != nil && t.finalizer.HandleError(err, w, r) {
		return
	}

	t.finalizer.HandleSuccess(w, res)

}

// TestRPCer invoke *ControllerJSONGen.TestRPCer using the request body as a json payload.
// TestRPCer Decodes r as json to invoke *Controller.TestRPCer.
// TestRPCer ...
func (t *ControllerHTTPGen) TestRPCer(w http.ResponseWriter, r *http.Request) {

	res, err := t.embed.TestRPCer(r)
	if err != nil && t.finalizer.HandleError(err, w, r) {
		return
	}

	t.finalizer.HandleSuccess(w, res)

}
```

# Recipes

#### Release the project

```sh
gump patch -d # check
gump patch # bump
```

# History

[CHANGELOG](CHANGELOG.md)

package main

import (
	"log"
	"net/http"

	httper "github.com/mh-cbon/httper/lib"
)

//go:generate lister vegetables_gen.go *Tomate:Tomates
//go:generate channeler tomate_chan_gen.go *Tomates:ChanTomates

//go:generate jsoner json_controller_gen.go *Controller:JSONController
//go:generate httper -mode gorilla http_vegetables_gen.go *JSONController:HTTPController

func main() {

	backend := NewChanTomates()
	backend.Push(&Tomate{Name: "red"})

	jsoner := NewJSONController(NewController(backend))
	httper := NewHTTPController(jsoner)

	// public views
	http.HandleFunc("/", httper.GetByID)

	/*
		curl -H "Accept: application/json" -H "Content-type: application/json"  http://localhost:8080/?id=0
	*/

	log.Fatal(http.ListenAndServe(":8080", nil))
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
	backend *ChanTomates
}

// NewController ...
func NewController(backend *ChanTomates) *Controller {
	return &Controller{
		backend: backend,
	}
}

// GetByID ...
func (t *Controller) GetByID(urlID int) *Tomate {
	return t.backend.Filter(FilterTomates.ByID(urlID)).First()
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

// TestRPCer ...
func (t *Controller) TestRPCer(id int) bool {
	return false
}

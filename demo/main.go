package main

import (
	"log"
	"net/http"
)

//go:generate lister vegetables_gen.go *Tomate:Tomates
//go:generate channeler tomate_chan_gen.go *Tomates:ChanTomates

//go:generate jsoner json_controller_gen.go *Controller:JSONController
//go:generate httper http_vegetables_gen.go *JSONController:HTTPController

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
	t.backend.Filter(func(v *Tomate) bool {
		if v.ID == urlID {
			v.Name = reqBody.Name
		}
		return true
	})
	return reqBody
}

// DeleteByID ...
func (t *Controller) DeleteByID(reqID int) bool {
	return t.backend.Remove(&Tomate{ID: reqID})
}

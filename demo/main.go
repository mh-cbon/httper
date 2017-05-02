package main

import (
	"log"
	"net/http"
)

// Tomate is about red vegetables to make famous italian food.
type Tomate struct {
	Name string
}

// GetID return the ID of the Tomate.
func (t Tomate) GetID() string {
	return t.Name
}

//go:generate lister vegetables_gen.go Tomate:Tomates
//go:generate jsoner json_vegetables_gen.go *Tomates:JSONTomates
//go:generate httper http_vegetables_gen.go *JSONTomates:HTTPTomates

func main() {

	backend := NewTomates()
	backend.Push(Tomate{Name: "red"})
	jsoner := NewJSONTomates(backend)
	httper := NewHTTPTomates(jsoner)

	// public views
	http.HandleFunc("/", httper.At)

	/*
		curl -H "Accept: application/json" -H "Content-type: application/json" -X POST -d ' {"i":0}'  http://localhost:8080/
	*/

	log.Fatal(http.ListenAndServe(":8080", nil))
}

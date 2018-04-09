// Copyright 2015 The Loadcat Authors. All rights reserved.

package ui

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"gopkg.in/mgo.v2/bson"

	"github.com/b-/loadcat/data"
	"github.com/b-/loadcat/feline"
	"github.com/b-/loadcat/templates"
)

var (
	logger = log.New(os.Stderr, "WARNING:", log.Llongfile|log.LstdFlags)
)

// ServeBalancerList - we received a request for the list of Balancers, so we serve that list of Balancers/
func ServeBalancerList(w http.ResponseWriter, r *http.Request) {
	bals, err := data.ListBalancers()
	if err != nil {
		panic(err)
	}

	err = templates.TplBalancerList.Execute(w, struct {
		Balancers []data.Balancer
	}{
		Balancers: bals,
	})
	if err != nil {
		panic(err)
	}
}

// ServeBalancerNewForm - we received the request to make a new Balancer, so we serve the form to fill out.
func ServeBalancerNewForm(w http.ResponseWriter, r *http.Request) {
	err := templates.TplBalancerNewForm.Execute(w, struct {
	}{})
	if err != nil {
		panic(err)
	}
}

// HandleBalancerCreate - Create a new balancer based on the POSTed form data.
func HandleBalancerCreate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	body := struct {
		Label string `schema:"label"`
	}{}
	err = schema.NewDecoder().Decode(&body, r.PostForm)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	bal := data.Balancer{}
	bal.Label = body.Label
	err = bal.Put()
	if err != nil {
		panic(err)
	}

	err = feline.Commit(&bal)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/balancers/"+bal.Id.Hex()+"/edit", http.StatusSeeOther)
}

// ServeBalancer - we received a request for a specific Balancer, so we serve the details on it.
func ServeBalancer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if !bson.IsObjectIdHex(vars["id"]) { // If the requested Balancer's ID is not a valid ID,
		http.Error(w, "Not Found", http.StatusNotFound) // serve an error 404.
		return
	}
	bal, err := data.GetBalancer(bson.ObjectIdHex(vars["id"]))
	if err != nil {
		//panic(err)
		logger.Output(2, err.Error())
	}

	err = templates.TplBalancerView.Execute(w, struct {
		Balancer *data.Balancer
	}{
		Balancer: bal,
	})
	if err != nil {
		//panic(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		logger.Output(2, err.Error())

	}
}

// ServeBalancerEditForm - we received a request to edit a specific balancer, so we serve the edit form
func ServeBalancerEditForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if !bson.IsObjectIdHex(vars["id"]) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	bal, err := data.GetBalancer(bson.ObjectIdHex(vars["id"]))
	if err != nil {
		panic(err)
	}

	err = templates.TplBalancerEditForm.Execute(w, struct {
		Balancer     *data.Balancer
		Protocols    []data.Protocol
		Algorithms   []data.Algorithm
		CipherSuites []data.CipherSuite
	}{
		Balancer:     bal,
		Protocols:    data.Protocols,
		Algorithms:   data.Algorithms,
		CipherSuites: data.CipherSuites,
	})
	if err != nil {
		panic(err)
	}
}

func HandleBalancerUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if !bson.IsObjectIdHex(vars["id"]) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	bal, err := data.GetBalancer(bson.ObjectIdHex(vars["id"]))
	if err != nil {
		panic(err)
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	body := struct {
		Label    string `schema:"label"`
		Settings struct {
			Hostname   string `schema:"hostname"`
			Port       int    `schema:"port"`
			Protocol   string `schema:"protocol"`
			Algorithm  string `schema:"algorithm"`
			SSLOptions struct {
				LetsEncrypt bool    `schema:letsencrypt`
				CipherSuite string  `schema:"cipher_suite"`
				Certificate *string `schema:"certificate"`
				PrivateKey  *string `schema:"private_key"`
			} `schema:"ssl_options"`
		} `schema:"settings"`
	}{}
	err = schema.NewDecoder().Decode(&body, r.PostForm)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	bal.Label = body.Label
	bal.Settings.Hostname = body.Settings.Hostname
	bal.Settings.Port = body.Settings.Port
	bal.Settings.Protocol = data.Protocol(body.Settings.Protocol)
	bal.Settings.Algorithm = data.Algorithm(body.Settings.Algorithm)
	if body.Settings.Protocol == "https" {
		bal.Settings.SSLOptions.LetsEncrypt = body.Settings.SSLOptions.LetsEncrypt
		bal.Settings.SSLOptions.CipherSuite = "recommended"
		if !body.Settings.SSLOptions.LetsEncrypt { //if body.Settings.SSLOptions.Certificate != nil {
			bal.Settings.SSLOptions.Certificate = []byte(*body.Settings.SSLOptions.Certificate)
			//}
			//if body.Settings.SSLOptions.PrivateKey != nil {
			bal.Settings.SSLOptions.PrivateKey = []byte(*body.Settings.SSLOptions.PrivateKey)
		}
	}
	err = bal.Put()
	if err != nil {
		panic(err)
	}

	err = feline.Commit(bal)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/balancers/"+bal.Id.Hex(), http.StatusSeeOther)
}

func init() {
	Router.NewRoute().
		Methods("GET").
		Path("/balancers").
		Handler(http.HandlerFunc(ServeBalancerList))
	Router.NewRoute().
		Methods("GET").
		Path("/balancers/new").
		Handler(http.HandlerFunc(ServeBalancerNewForm))
	Router.NewRoute().
		Methods("POST").
		Path("/balancers/new").
		Handler(http.HandlerFunc(HandleBalancerCreate))
	Router.NewRoute().
		Methods("GET").
		Path("/balancers/{id}").
		Handler(http.HandlerFunc(ServeBalancer))
	Router.NewRoute().
		Methods("GET").
		Path("/balancers/{id}/edit").
		Handler(http.HandlerFunc(ServeBalancerEditForm))
	Router.NewRoute().
		Methods("POST").
		Path("/balancers/{id}/edit").
		Handler(http.HandlerFunc(HandleBalancerUpdate))
}

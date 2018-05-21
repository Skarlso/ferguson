package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

// PostJob accepts a yaml file to parse which contains steps to do.
func PostJob(w http.ResponseWriter, r *http.Request) {
	j := Job{}
	if r.Body == nil {
		fmt.Fprintln(w, "empty body")
		return
	}
	defer r.Body.Close()
	buff, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintln(w, "error while reading body: ", err)
		return
	}
	err = yaml.Unmarshal(buff, &j)
	if err != nil {
		fmt.Fprintln(w, "error while unmarshalling body: ", err)
		return
	}
	j.Parse()
	fmt.Println("got translated commands: ", j.Translated)
}

func main() {
	// Loading Server plugins before the server starts
	Load("plugins/git.lua")
	Load("plugins/bash.lua")

	// Calling plugin
	// val, _ := Call("func", 6)
	// log.Println("Got from the script: ", val)

	server := new(Server)
	server.populateAgentMap()
	go server.listen()
	health := func() {
		for {
			server.sendHealthCheckToAgents()
			time.Sleep(30 * time.Second)
		}
	}
	go health()

	router := mux.NewRouter()
	router.HandleFunc("/jobs/add", PostJob).Methods("POST")
	log.Fatal(http.ListenAndServe(":8000", router))
}

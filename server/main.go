package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
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

func loadPlugins() {
	files, err := ioutil.ReadDir("./plugins")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		Load(filepath.Join("plugins", file.Name()))
		log.Println("loaded plugin: ", (file.Name()))
	}
}

func main() {
	loadPlugins()

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

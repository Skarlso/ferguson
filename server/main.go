package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

// TODO: Need a way to persist this count
var jobCount int

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
	// server.SendToNoneBusyWorker(j.Translated)
	// CREATE a Job! The job will have an agent assigned
	// As long as the agent is assigned to a Job it's busy.
	// And the job will take care of the commands that it needs to execute on that agent.

	server.executeViaSSH(j.Translated)
}

// GetJob will attach to the log output of the job with number ID.
func GetJob(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	id, err := strconv.Atoi(v.Get("id"))
	if err != nil {
		fmt.Fprintf(w, "cannot convert to number: '%s'", id)
	}
	fmt.Fprintf(w, "looking up job number: '%d'", id)
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

var server Server

func main() {
	loadPlugins()

	server.populateAgentMap()
	go server.sshListen()

	health := func() {
		for {
			server.sendHealthCheckToSSHAgents()
			time.Sleep(30 * time.Second)
		}
	}
	go health()

	router := mux.NewRouter()
	router.HandleFunc("/jobs/add", PostJob).Methods("POST")
	router.HandleFunc("/jobs/{id:[0-9]+}", GetJob).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}

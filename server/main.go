package main

import (
	"encoding/json"
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

var jobQueue [][]string

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
	ssha := server.getIdleWorker()
	if ssha == nil {
		jobQueue = append(jobQueue, j.Translated)
		fmt.Fprintln(w, "job added to queue")
		return
	}
	ssha.Busy = true
	rj := RunningJob{
		Agent: ssha,
		Count: jobCount,
	}
	go rj.executeViaSSH(j.Translated)
	jobCount++
	saveJobCount(jobCount)
}

// GetJob will attach to the log output of the job with number ID.
func GetJob(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id, err := strconv.Atoi(v["id"])
	if err != nil {
		fmt.Fprintf(w, "cannot convert to number: '%v'. err: %v", id, err)
		return
	}
	l, err := ioutil.ReadFile(filepath.Join("logs", v["id"]+".log"))
	if err != nil {
		fmt.Fprintln(w, "couldn't read log file: ", err)
		return
	}
	fmt.Fprintln(w, string(l))
}

// QueueJob is the upper wrapper of a job json
type QueueJob struct {
	Job []string `json:"job"`
}

// QueueList is a list of jobs that are waiting in the queue to be processed
type QueueList struct {
	Jobs []QueueJob `json:"jobs"`
}

// ListQueue returns any items that are waiting in line to be processed.
func ListQueue(w http.ResponseWriter, r *http.Request) {
	qj := new(QueueList)
	for _, q := range jobQueue {
		j := QueueJob{
			Job: q,
		}
		qj.Jobs = append(qj.Jobs, j)
	}
	b, _ := json.Marshal(qj)
	fmt.Fprintf(w, "%s", string(b))
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

func init() {
	jobCount = loadJobCount()
	jobQueue = make([][]string, 0)
}

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
	router.HandleFunc("/job/{id:[0-9]+}", GetJob).Methods("GET")
	router.HandleFunc("/jobs/listQueue", ListQueue).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}

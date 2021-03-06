package main

import (
	"io/ioutil"
	"log"
	"strconv"
	"sync"

	"github.com/yuin/gopher-lua"
)

// Job is a single job.
type Job struct {
	Name       string        `yaml:"name"`
	Stages     []interface{} `yaml:"steps"`
	Translated []string
}

// RunningJob will take care of a running job and all it's affiliations.
type RunningJob struct {
	// Must be mutex.Lock()-ed so it's not repeated
	Count int
	// The Agent that is running this job
	Agent *SSHAgent
	// The outcome of the running job
	Status bool
}

// Parse will translate the stages into executable bash scripts.
func (j *Job) Parse() {
	translated := make([]string, 0)
	// Parse the data as key = value pairs and pass it over to the plugin
	// the plugin has to deal with applying the data internally.
	for _, steps := range j.Stages {
		for k, v := range steps.(map[interface{}]interface{}) {
			params := L.NewTable()
			for _, p := range v.([]interface{}) {
				for key, value := range p.(map[interface{}]interface{}) {
					params.RawSetString(key.(string), lua.LString(value.(string)))
				}
			}
			cmd, err := Call(k.(string), *params)
			if err != nil {
				log.Println("error while calling function: ", err)
				return
			}
			translated = append(translated, cmd.String())
		}
	}
	j.Translated = translated
}

func (jr *RunningJob) executeViaSSH(cmds []string) {
	err := jr.Agent.dialAndSend(cmds)
	if err != nil {
		jr.Status = false
	}
	jr.Status = true
	jr.Agent.Busy = false
	// signal the queue clearer that an agent got free
	go queueClearer()
}

// This also needs to be running periodically in case there are no agents.
func queueClearer() {
	if len(jobQueue) == 0 {
		return
	}
	ssha := server.getIdleWorker()
	if ssha == nil {
		return
	}
	var job []string
	ssha.Busy = true
	job, jobQueue = jobQueue[0], jobQueue[1:]
	rj := RunningJob{
		Agent: ssha,
		Count: jobCount,
	}
	jobCount++
	go rj.executeViaSSH(job)
	saveJobCount(jobCount)
}

func saveJobCount(jc int) {
	m := sync.Mutex{}
	m.Lock()
	defer m.Unlock()
	c := strconv.Itoa(jobCount)
	err := ioutil.WriteFile(".job_count", []byte(c), 0644)
	if err != nil {
		log.Println("error while saving version: ", err)
	}
}

func loadJobCount() int {
	b, _ := ioutil.ReadFile(".job_count")
	c, _ := strconv.Atoi(string(b))
	return c
}

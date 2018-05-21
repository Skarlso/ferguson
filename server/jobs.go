package main

import (
	"log"

	"github.com/yuin/gopher-lua"
)

// Job is a single job.
type Job struct {
	Name       string        `yaml:"name"`
	Stages     []interface{} `yaml:"steps"`
	Translated []string
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

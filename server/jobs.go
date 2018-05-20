package main

import "fmt"

// Job is a single job.
type Job struct {
	Name       string                            `yaml:"name"`
	Stages     map[string]map[string]interface{} `yaml:"stages"`
	Translated map[string][]string
}

// Parse will translate the stages into executable bash scripts.
func (j *Job) Parse() {
	translated := make(map[string][]string, 0)
	for k, v := range j.Stages {
		commands := make([]string, 0)
		for p, cmd := range v {
			params := make([]interface{}, 0)
			for _, p := range cmd.([]interface{}) {
				for _, param := range p.(map[interface{}]interface{}) {
					params = append(params, param)
				}
			}
			t, err := Call(p, params...)
			if err != nil {
				fmt.Printf("problem during transalting step '%s'. error: %v\n", t, err)
			}
			commands = append(commands, t.String())
		}
		translated[k] = commands
	}
	j.Translated = translated
}

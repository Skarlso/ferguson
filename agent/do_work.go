package main

import "log"

// DoWork performs the work that this agent received.
// TODO: Pass over the connection so it can signal master that it's finished with it.
func DoWork(work [][]byte) {
	for _, w := range work {
		log.Println("doing work: ", string(w))
	}
}

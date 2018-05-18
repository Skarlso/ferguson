package main

import "log"

func main() {

	// Loading Server plugins before the server starts
	Load("plugins/simple.lua")

	// Calling plugin
	val, _ := Call("func")
	log.Println("Got from the script: ", val)

	server := new(Server)
	server.populateAgentMap()
	go server.listen()
	for {
		server.sendHealthCheckToAgents()
	}
}

package main

func main() {

	// Loading Server plugins before the server starts
	Load("plugins/simple.lua", "func")

	server := new(Server)
	server.populateAgentMap()
	go server.listen()
	for {
		server.sendHealthCheckToAgents()
	}
}

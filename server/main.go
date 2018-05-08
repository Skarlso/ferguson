package main

func main() {
	server := new(Server)
	server.populateAgentMap()
	go server.listen()
	for {
		server.sendHealthCheckToAgents()
	}
}

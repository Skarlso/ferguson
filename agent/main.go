package main

// Agent is an object wrapping an agent which connects to a server.
type Agent struct {
}

func main() {
	agent := new(Agent)
	agent.connectViaSSH()
}

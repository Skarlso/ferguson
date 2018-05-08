package main

import (
	"crypto/tls"
	"io"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

// Agent is an object wrapping an agent which connects to a server.
type Agent struct {
}

func (a *Agent) connect() {
	serverAddr := os.Getenv("SERVER_ADDRESS")
	cert, err := tls.LoadX509KeyPair("certs/agent.pem", "certs/agent.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", serverAddr, &config)
	if err != nil {
		log.Fatalf("agent: dial: %s", err)
	}
	log.Println("agent: connected to: ", conn.RemoteAddr())

	state := conn.ConnectionState()

	log.Println("agent: handshake: ", state.HandshakeComplete)
	log.Println("agent: mutual: ", state.NegotiatedProtocolIsMutual)

	message := "Agent connected. Ready to accept work.\n"
	_, err = io.WriteString(conn, message)
	if err != nil {
		log.Fatalf("agent: write: %s", err)
	}
	defer conn.Close()
	buf := make([]byte, 512)
	for {
		log.Print("server: conn: waiting")
		n, err := conn.Read(buf)
		if err != nil {
			if err != nil {
				log.Printf("server: conn: read: %s", err)
			}
			break
		}
		log.Printf("server: conn: echo %q\n", string(buf[:n]))
	}
	log.Println("server: conn: closed")
}

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
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
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", serverAddr, &config)
	if err != nil {
		log.Fatalf("agent: dial: %s", err)
	}
	defer conn.Close()
	log.Println("agent: connected to: ", conn.RemoteAddr())

	state := conn.ConnectionState()
	for _, v := range state.PeerCertificates {
		fmt.Println(x509.MarshalPKIXPublicKey(v.PublicKey))
		fmt.Println(v.Subject)
	}
	log.Println("agent: handshake: ", state.HandshakeComplete)
	log.Println("agent: mutual: ", state.NegotiatedProtocolIsMutual)

	message := "Hello\n"
	n, err := io.WriteString(conn, message)
	if err != nil {
		log.Fatalf("agent: write: %s", err)
	}
	log.Printf("agent: wrote %q (%d bytes)", message, n)

	reply := make([]byte, 256)
	n, err = conn.Read(reply)
	log.Printf("agent: read %q (%d bytes)", string(reply[:n]), n)
	log.Print("agent: exiting")
}

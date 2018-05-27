package main

import (
	"io/ioutil"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/crypto/ssh"
)

func (a *Agent) connectViaSSH() {
	serverAddr := os.Getenv("SERVER_ADDRESS")
	// var hostKey ssh.PublicKey

	key, err := ioutil.ReadFile("/Users/hannibal/.ssh/id_rsa")
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: "hannibal",
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		// Normally this would be ssh.FixedHostKey(hostKey),
		// In which case I would have to handle adding an unkown hostkey
		// to the list of `known_hosts`.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to the remote server and perform the SSH handshake.
	client, err := ssh.Dial("tcp", serverAddr, config)
	if err != nil {
		log.Fatalf("unable to connect: %v", err)
	}
	log.Println("successfully registered at master. awaiting commands...")
	defer client.Close()
}

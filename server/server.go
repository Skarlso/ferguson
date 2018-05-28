package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/go-redis/redis"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/crypto/ssh"
)

var client *redis.Client

func init() {
	redisAddr := os.Getenv("REDIS_ADDRESS")
	client = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatal("failed to connect to redis server on: ", redisAddr)
	}
}

// Server defines a server object which has various capabilities that a server requires.
type Server struct {
	agents sync.Map
	count  int
}

// SSHAgent defines an agent which can be ssh'ed to.
// Handles connections based on hostname.
type SSHAgent struct {
	Busy     bool
	Hostname string
}

var authorizedKeysMap = make(map[string]bool, 0)

func loadAuthorizedKeys() {
	// Public key authentication is done by comparing
	// the public key of a received connection
	// with the entries in the authorized_keys file.
	authorizedKeysBytes, err := ioutil.ReadFile("/Users/hannibal/.ssh/authorized_keys")
	if err != nil {
		log.Fatalf("Failed to load authorized_keys, err: %v", err)
	}

	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			log.Fatal(err)
		}

		authorizedKeysMap[string(pubKey.Marshal())] = true
		authorizedKeysBytes = rest
	}
}

func authorizeConnection(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
	if authorizedKeysMap[string(pubKey.Marshal())] {
		return &ssh.Permissions{
			// Record the public key used for authentication.
			Extensions: map[string]string{
				"pubkey-fp": ssh.FingerprintSHA256(pubKey),
			},
		}, nil
	}
	return nil, fmt.Errorf("unknown public key for %q", c.User())
}

func (s *Server) sshListen() {
	loadAuthorizedKeys()
	config := ssh.ServerConfig{
		PublicKeyCallback: authorizeConnection,
	}
	privateBytes, err := ioutil.ReadFile("/Users/hannibal/.ssh/id_rsa")
	if err != nil {
		log.Fatal("Failed to load private key: ", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}

	config.AddHostKey(private)
	listeningAddr := os.Getenv("LISTENING_ADDRESS")
	socket, err := net.Listen("tcp", listeningAddr)
	if err != nil {
		panic(err)
	}
	log.Println("listening for ssh connections...")
	for {
		conn, err := socket.Accept()
		if err != nil {
			panic(err)
		}

		// From a standard TCP connection to an encrypted SSH connection
		sshConn, _, _, err := ssh.NewServerConn(conn, &config)
		if err != nil {
			panic(err)
		}

		log.Println("Connection from", sshConn.RemoteAddr())
		ssha := SSHAgent{
			Hostname: sshConn.RemoteAddr().String(),
			Busy:     false,
		}
		s.saveSSHClient(&ssha)
		sshConn.Close()
	}
}

func (ssha *SSHAgent) dialAndSend(commands []string) error {
	sshPort := os.Getenv("SSH_PORT")
	sshUser := os.Getenv("SSH_USER")
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
		User: sshUser,
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		// Normally this would be ssh.FixedHostKey(hostKey),
		// In which case I would have to handle adding an unkown hostkey
		// to the list of `known_hosts`.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	hostName := ssha.Hostname[:strings.IndexByte(ssha.Hostname, ':')]
	// Connect to the remote server and perform the SSH handshake.
	client, err := ssh.Dial("tcp", hostName+":"+sshPort, config)
	if err != nil {
		log.Fatalf("unable to connect: %v", err)
	}
	defer client.Close()

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	for _, cmd := range commands {
		log.Println("running: ", cmd)
		if err := session.Run(cmd); err != nil {
			return err
		}
		log.Println(b.String())
	}
	return nil
}

func (s *Server) executeViaSSH(commands []string) {
	if s.count < 1 {
		return
	}
	f := func(key, value interface{}) bool {
		ssha := value.(*SSHAgent)
		log.Println("sending work to host: ", ssha.Hostname)
		ssha.dialAndSend(commands)
		return false
	}
	s.agents.Range(f)
}

// Load in the connection detail of all agenst from redis and
// populate the agents map for this server with nil connections.
func (s *Server) populateAgentMap() {

}

// Create the actual connection to all agents in go routines and save
// that connection in the agents map for further use.
func (s *Server) createConnectionsToAgents() {

}

// Send a general ping to the agents recording response time in ms.
// This is using sending of work right now, but it will use a ping.
// if the ping comes back as errored, we get rid of the worker.
func (s *Server) sendHealthCheckToSSHAgents() {
	f := func(key, value interface{}) bool {
		ssha := value.(*SSHAgent)
		if err := ssha.dialAndSend([]string{"echo"}); err != nil {
			log.Println("deleting host: ", key.(string))
			s.agents.Delete(key)
			s.count--
		}
		return true
	}
	s.agents.Range(f)
}

// Save / Update an agent record in redis?
func (s *Server) saveSSHClient(a *SSHAgent) {
	s.agents.Store(a.Hostname, a)
	s.count++
}

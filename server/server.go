package main

import (
	"crypto/rand"
	"crypto/tls"
	"log"
	"os"
	"sync"

	"github.com/go-redis/redis"
	_ "github.com/joho/godotenv/autoload"
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
}

func (s *Server) listen() {
	listeningAddr := os.Getenv("LISTENING_ADDRESS")
	cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	config.Rand = rand.Reader
	listener, err := tls.Listen("tcp", listeningAddr, &config)
	if err != nil {
		log.Fatalf("server: listen: %s", err)
	}
	log.Print("server: listening")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}
		log.Printf("server: accepted from %s", conn.RemoteAddr())
		s.saveClient(conn.(*tls.Conn))
	}
}

func (s *Server) connectToAgentAddress(addrs string) *tls.Conn {
	cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", addrs, &config)
	if err != nil {
		log.Print("no agent yet: ", err)
		return nil
	}
	log.Println("agent: connected to: ", conn.RemoteAddr())

	state := conn.ConnectionState()

	log.Println("agent: handshake: ", state.HandshakeComplete)
	log.Println("agent: mutual: ", state.NegotiatedProtocolIsMutual)
	return conn
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
func (s *Server) sendHealthCheckToAgents() {
	f := func(key, value interface{}) bool {
		if value == nil {
			return false
		}
		if err := s.sendWork(value.(*tls.Conn)); err != nil {
			s.agents.Delete(key)
		}
		return true
	}
	s.agents.Range(f)
}

// Save / Update an agent record in redis?
func (s *Server) saveClient(conn *tls.Conn) {
	s.agents.Store(conn.RemoteAddr(), conn)
}

func (s *Server) sendWork(conn *tls.Conn) error {
	work := []byte("This is important work.")
	_, err := conn.Write(work)
	if err != nil {
		log.Printf("%s went away, deleting from agents.", conn.RemoteAddr())
		return err
	}
	// log.Printf("server: conn: wrote %d bytes", n)
	return nil
}

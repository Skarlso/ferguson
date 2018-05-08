package main

import (
	"crypto/rand"
	"crypto/tls"
	"log"
	"net"
	"os"

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
		defer conn.Close()
		log.Printf("server: accepted from %s", conn.RemoteAddr())
		// tlscon, ok := conn.(*tls.Conn)
		// if ok {
		// 	log.Print("ok=true")
		// 	state := tlscon.ConnectionState()
		// 	for _, v := range state.PeerCertificates {
		// 		log.Print(x509.MarshalPKIXPublicKey(v.PublicKey))
		// 	}
		// }
		go saveClient(conn)
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

func saveClient(conn net.Conn) {
	defer conn.Close()
	client.Set("agent1", conn.RemoteAddr(), 0)
}

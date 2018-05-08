package main

import (
	"log"
	"net"
)

func (s *Server) sendWork(conn net.Conn) {
	defer conn.Close()
	work := []byte("This is important work.")
	n, err := conn.Write(work)
	if err != nil {
		log.Fatal("server: write: ", err)
	}
	log.Printf("server: conn: wrote %d bytes", n)
}

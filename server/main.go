package main

import (
	"github.com/go-redis/redis"
)

func main() {
	server := new(Server)
	go server.listen()
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	for {
		agent, _ := client.Get("agent1").Result()
		conn := server.connectToAgentAddress(agent)
		if conn != nil {
			server.sendWork(conn)
		}
	}
}

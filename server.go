package main

import (
	"fmt"
	"log"
	"net"
	rpc "net/rpc"
	"sync"
	"time"
)

type ChatServer struct{}

type Client struct {
	ID   int
	Name string
	Ch   chan string
}

var (
	clients   = make(map[int]*Client)
	clientMux sync.Mutex
	nextID    = 1
)

type RegisterArgs struct {
	UserName string
}

type RegisterReply struct {
	Success bool
	ID      int
}

type SendMessageArgs struct {
	UserID  int
	Content string
}

type ReceiveArgs struct {
	UserID int
}

func (s *ChatServer) Register(args *RegisterArgs, reply *RegisterReply) error {
	clientMux.Lock()
	defer clientMux.Unlock()

	id := nextID
	nextID++

	newClient := &Client{
		ID:   id,
		Name: args.UserName,
		Ch:   make(chan string, 20),
	}

	clients[id] = newClient

	joinMsg := fmt.Sprintf("User [%d] joined", id)
	for uid, cl := range clients {
		if uid != id {
			select {
			case cl.Ch <- joinMsg:
			default:
				log.Printf("Warning: channel full for user %d", uid)
			}
		}
	}

	reply.Success = true
	reply.ID = id

	log.Printf("User [%d] (%s) registered. Total clients: %d", id, args.UserName, len(clients))
	return nil
}

func (s *ChatServer) SendMessage(args *SendMessageArgs, reply *bool) error {
	clientMux.Lock()
	sender, ok := clients[args.UserID]
	clientMux.Unlock()

	if !ok {
		*reply = false
		return fmt.Errorf("user [%d] not registered", args.UserID)
	}

	msg := fmt.Sprintf("[%s] User [%d] (%s): %s",
		time.Now().Format("15:04:05"),
		sender.ID,
		sender.Name,
		args.Content,
	)

	clientMux.Lock()
	defer clientMux.Unlock()

	sentCount := 0
	for uid, cl := range clients {
		if uid == args.UserID {
			continue
		}

		select {
		case cl.Ch <- msg:
			sentCount++
		default:
			log.Printf("Warning: channel full for user %d", uid)
		}
	}

	log.Printf("User [%d] sent message to %d clients", args.UserID, sentCount)
	*reply = true
	return nil
}

func (s *ChatServer) Receive(args *ReceiveArgs, msg *string) error {
	clientMux.Lock()
	c, ok := clients[args.UserID]
	clientMux.Unlock()

	if !ok {
		return fmt.Errorf("client [%d] not registered", args.UserID)
	}

	*msg = <-c.Ch
	return nil
}

func (s *ChatServer) Disconnect(userID int, reply *bool) error {
	clientMux.Lock()
	defer clientMux.Unlock()

	client, ok := clients[userID]
	if !ok {
		*reply = false
		return fmt.Errorf("user [%d] not found", userID)
	}

	close(client.Ch)
	delete(clients, userID)

	disconnectMsg := fmt.Sprintf("User [%d] left", userID)
	for _, cl := range clients {
		select {
		case cl.Ch <- disconnectMsg:
		default:
		}
	}

	*reply = true
	log.Printf("User [%d] (%s) disconnected. Remaining clients: %d",
		userID, client.Name, len(clients))
	return nil
}

func main() {
	server := new(ChatServer)
	rpc.Register(server)

	addr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:42586")
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
	defer listener.Close()

	fmt.Println("====================================")
	fmt.Println("Real-time Chat Server Started")
	fmt.Println("====================================")
	fmt.Println("Listening on: tcp://0.0.0.0:42586")
	fmt.Println("Waiting for clients...")
	fmt.Println("====================================")

	rpc.Accept(listener)
}
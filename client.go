package main

import (
	"bufio"
	"fmt"
	"log"
	rpc "net/rpc"
	"os"
	"strings"
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

var (
	client   *rpc.Client
	userID   int
	userName string
)

func main() {
	var err error
	client, err = rpc.Dial("tcp", "0.0.0.0:42586")
	if err != nil {
		log.Fatal("Error connecting to server:", err)
	}
	defer client.Close()

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter your name: ")
	scanner.Scan()
	userName = strings.TrimSpace(scanner.Text())
	if userName == "" {
		userName = "Anonymous"
	}

	var regReply RegisterReply
	err = client.Call("ChatServer.Register", &RegisterArgs{UserName: userName}, &regReply)
	if err != nil || !regReply.Success {
		log.Fatal("Failed to register with server:", err)
	}

	userID = regReply.ID

	fmt.Println("====================================")
	fmt.Printf("Connected as User [%d] (%s)\n", userID, userName)
	fmt.Println("====================================")
	fmt.Println("Type 'exit' to quit")
	fmt.Println("====================================\n")

	go receiveMessages()

	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			break
		}

		text := strings.TrimSpace(scanner.Text())

		if text == "" {
			continue
		}

		if text == "exit" {
			fmt.Println("Goodbye!")
			disconnect()
			return
		}

		args := SendMessageArgs{
			UserID:  userID,
			Content: text,
		}
		var ok bool
		err := client.Call("ChatServer.SendMessage", &args, &ok)
		if err != nil {
			fmt.Printf("Send failed: %v\n", err)
		}
	}
}

func receiveMessages() {
	for {
		var msg string

		err := client.Call("ChatServer.Receive", &ReceiveArgs{UserID: userID}, &msg)
		if err != nil {
			fmt.Printf("\nDisconnected from server: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n%s\n> ", msg)
	}
}

func disconnect() {
	var ok bool
	client.Call("ChatServer.Disconnect", userID, &ok)
}
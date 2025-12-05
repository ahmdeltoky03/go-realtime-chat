# Go Real-time Chat System

A concurrent chat application using Go RPC, Goroutines, and Channels with real-time message broadcasting.

## Features

- Real-time message broadcasting
- User join/leave notifications
- No self-echo (sender doesn't receive own messages)
- Thread-safe using Mutex
- Concurrent operations with goroutines and channels
- Graceful disconnect handling

## Requirements

- Go 1.16 or higher

## Installation

```bash
git clone https://github.com/YOUR_USERNAME/go-realtime-chat.git
cd go-realtime-chat
go mod tidy
```

## Usage

Terminal 1 - Start Server:
```bash
go run server.go
```

Terminal 2 - Start Client 1:
```bash
go run client.go
```
Enter name: Alice

Terminal 3 - Start Client 2:
```bash
go run client.go
```
Enter name: Bob

Now Alice and Bob can chat. Messages are broadcast to all clients except the sender.

## How It Works

Server:
- Register: Adds client to system, notifies others
- SendMessage: Broadcasts message to all clients except sender
- Receive: Blocking call for client to get next message
- Disconnect: Removes client, notifies others

Client:
- Main thread: Reads user input, sends via RPC
- Goroutine: Listens for incoming messages

## Architecture

Key components:
- clients map[int]*Client: Stores all connected clients
- clientMux: Mutex protecting concurrent access
- Ch: Buffered channel for each client (size 20)
- nextID: Auto-increment counter for user IDs

## Project Structure

```
go-realtime-chat/
├── server.go
├── client.go
├── go.mod
├── .gitignore
└── README.md
```

## Testing

1. Start server: go run server.go
2. Open 2+ terminals
3. Run client in each: go run client.go
4. Type messages and press Enter
5. Type 'exit' to quit

Messages from one client are sent to all others with no self-echo.
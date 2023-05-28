package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	chat "github.com/mattmazer1/grpc-server-comms/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var client chat.ChatAppClient
var wait *sync.WaitGroup

func init() {
	wait = &sync.WaitGroup{}
}

func connect(user *chat.User) error {
	var streamError error

	stream, err := client.CreateStream(context.Background(), &chat.Connect{
		User:   user,
		Active: true,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}

	wait.Add(1)
	go func(stream chat.ChatApp_CreateStreamClient) {
		defer wait.Done()

		for {
			msg, err := stream.Recv()
			if err != nil {
				streamError = fmt.Errorf("failed to receive message: %v", err)
				break
			}

			fmt.Printf("%v: %s: %v\n", msg.Id, msg.Message, msg.Time)
		}

	}(stream)
	return streamError
}

func main() {
	timestamp := time.Now()
	done := make(chan int)

	name := flag.String("Name", "Anonymous", "Name of user")
	flag.Parse()

	id := sha256.Sum256([]byte(timestamp.String() + *name))

	creds := insecure.NewCredentials()
	conn, err := grpc.Dial("localhost:9000", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("failed to connect to service: %v", err)
	}

	client = chat.NewChatAppClient(conn)
	user := &chat.User{
		Id:   hex.EncodeToString(id[:]),
		Name: *name,
	}

	connect(user)

	wait.Add(1)

	go func() {
		defer wait.Done()
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			msg := &chat.Message{
				Id:      user.Id,
				Message: scanner.Text(),
				Time:    timestamp.String(),
			}

			_, err := client.BroadcastMessage(context.Background(), msg)
			if err != nil {
				fmt.Printf("failed to send message: %v", err)
				break
			}
		}
	}()

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
}

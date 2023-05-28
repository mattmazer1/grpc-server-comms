package main

import (
	"context"
	"log"
	"net"
	"os"
	"sync"

	chat "github.com/mattmazer1/grpc-chat-app/proto"
	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
)

var grpcLog glog.LoggerV2

func init() {
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

type Connection struct {
	stream chat.ChatApp_CreateStreamServer
	id     string
	active bool
	err    chan error
}

type Server struct {
	chat.UnimplementedChatAppServer
	Connection []*Connection
}

func (c *Server) CreateStream(connect *chat.Connect, stream chat.ChatApp_CreateStreamServer) error {
	conn := &Connection{
		stream: stream,
		id:     connect.User.Id,
		active: true,
		err:    make(chan error),
	}

	c.Connection = append(c.Connection, conn)

	return <-conn.err
}

func (c *Server) BroadcastMessage(ctx context.Context, msg *chat.Message) (*chat.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)

	for _, conn := range c.Connection {
		log.Println(conn.id)
		wait.Add(1)

		go func(msg *chat.Message, conn *Connection) {
			defer wait.Done()

			if conn.active {
				err := conn.stream.Send(msg)
				grpcLog.Info("Sending message to: ", conn.id)

				if err != nil {
					grpcLog.Errorf("Error with stream %v. Error: %v", conn.stream, err)
					conn.active = false
					conn.err <- err
				}
			}
		}(msg, conn)
	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
	return &chat.Close{}, nil
}

func main() {

	connections := []*Connection{}
	s := &Server{Connection: connections}

	grpcServer := grpc.NewServer()

	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen on port 9000: %v", err)
	}

	grpcLog.Info("Starting server on port 9000")

	chat.RegisterChatAppServer(grpcServer, s)
	grpcServer.Serve(lis)
}

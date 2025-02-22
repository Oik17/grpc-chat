package handlers

import (
	"context"
	"fmt"
	"sync"

	proto "github.com/Oik17/gRPC-chat/gen"
)

type Connection struct {
	proto.UnimplementedBroadcastServer
	stream proto.Broadcast_CreateStreamServer
	id     string
	active bool
	error  chan error
}

type Pool struct {
	proto.UnimplementedBroadcastServer
	Connection []*Connection
}

func (p *Pool) CreateStream(pconn *proto.Connect, stream proto.Broadcast_CreateStreamServer) error {
	if pconn == nil || pconn.User == nil {
		fmt.Println("Error: Received nil pconn or pconn.User")
		return fmt.Errorf("invalid connection request: user information is missing")
	}

	conn := &Connection{
		stream: stream,
		id:     pconn.User.Id,
		active: true,
		error:  make(chan error),
	}

	if p.Connection == nil {
		p.Connection = []*Connection{}
	}
	p.Connection = append(p.Connection, conn)

	fmt.Printf("User %v connected successfully\n", conn.id)

	return nil
}

func (s *Pool) BroadcastMessage(ctx context.Context, msg *proto.Message) (*proto.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan struct{}) 

	for _, conn := range s.Connection {
		wait.Add(1)

		go func(msg *proto.Message, conn *Connection) {
			defer wait.Done()

			if conn == nil {
				fmt.Println("Skipping nil connection")
				return
			}

			if conn.active {
				err := conn.stream.Send(msg)
				fmt.Printf("Sending message to: %v from %v\n", conn.id, msg.Id)

				if err != nil {
					fmt.Printf("Error with Stream: %v - Error: %v\n", conn.stream, err)
					conn.active = false
					select {
					case conn.error <- err:
					default:
						fmt.Println("Error channel full, dropping error")
					}
				}
			}
		}(msg, conn)
	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
	return &proto.Close{}, nil
}

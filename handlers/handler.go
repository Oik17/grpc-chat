package controllers

import (
	proto "github.com/Oik17/gRPC-chat/gen"
)

type Connection struct {
	proto.UnimplementedBroadcastServer
	stream proto.Broadcast_CreateStreamServer
	id     string
	active bool
	error  chan error
}

type Pool struct{
	proto.UnimplementedBroadcastServer
	Connection []*Connection
}

func (p *Pool) CreateStream(pconn *proto.Connect, stream proto.Broadcast_CreateStreamServer) error{
	conn:=&Connection{
		stream: stream,
		id: pconn.User.Id,
		active:true,
		error: make(chan error),
	}
	p.Connection=append(p.Connection, conn)
	return <-conn.error
}
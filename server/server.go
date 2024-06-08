package server

import (
	"fmt"
	"net"
	"time"

	"github.com/poolpOrg/go-agentbuilder/protocol"
)

type Session struct {
	protocol *protocol.Protocol
}

func NewSession(conn net.Conn) *Session {
	return &Session{
		protocol: protocol.NewProtocol(conn),
	}
}

func (s *Session) LocalAddr() net.Addr {
	return s.protocol.LocalAddr()
}

func (s *Session) RemoteAddr() net.Addr {
	return s.protocol.RemoteAddr()
}

func (s *Session) Latency() time.Duration {
	return s.protocol.Latency()
}

func (s *Session) Request(payload interface{}) error {
	return s.protocol.Request(payload)
}

func (s *Session) Query(payload interface{}, responseHandler func(interface{}) error) error {
	return s.protocol.Query(payload, responseHandler)
}

type Server struct {
	localAddr string
	listener  net.Listener
	exit      chan struct{}
}

func NewServer(address string) *Server {
	return &Server{
		localAddr: address,
		exit:      make(chan struct{}),
	}
}

func (s *Server) ListenAndServe(handler func(*Session, <-chan protocol.Packet) error) error {
	l, err := net.Listen("tcp", s.localAddr)
	if err != nil {
		return err
	}
	defer l.Close()

	s.listener = l
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go func() {
			defer func() {
				fmt.Printf("%s : disconnected\n", conn.RemoteAddr())
				conn.Close()
			}()

			fmt.Printf("%s : connected\n", conn.RemoteAddr())
			session := NewSession(conn)

			if err := handler(session, session.protocol.Incoming()); err != nil {
				fmt.Printf("%s : error: %s\n", conn.RemoteAddr(), err)
			}
		}()
	}
}

package server

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/poolpOrg/go-agentbuilder/protocol"
)

type Session struct {
	sessionID uuid.UUID
	protocol  *protocol.Protocol
}

func NewSession(conn net.Conn) *Session {
	return &Session{
		sessionID: uuid.Must(uuid.NewRandom()),
		protocol:  protocol.NewProtocol(conn),
	}
}

func (s *Session) SessionID() uuid.UUID {
	return s.sessionID
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

	sessions      map[uuid.UUID]*Session
	sessionsMutex sync.RWMutex

	exit chan struct{}
}

func NewServer(address string) *Server {
	return &Server{
		localAddr: address,
		exit:      make(chan struct{}),
	}
}

func (s *Server) Sessions() []*Session {
	s.sessionsMutex.RLock()
	defer s.sessionsMutex.RUnlock()

	sessions := make([]*Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
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
			fmt.Printf("%s : connected\n", conn.RemoteAddr())
			session := NewSession(conn)
			s.sessionsMutex.Lock()
			s.sessions[session.SessionID()] = session
			s.sessionsMutex.Unlock()

			defer func() {
				fmt.Printf("%s : disconnected\n", conn.RemoteAddr())
				conn.Close()
				s.sessionsMutex.Lock()
				delete(s.sessions, session.SessionID())
				s.sessionsMutex.Unlock()
			}()

			if err := handler(session, session.protocol.Incoming()); err != nil {
				fmt.Printf("%s : error: %s\n", conn.RemoteAddr(), err)
			}
		}()
	}
}

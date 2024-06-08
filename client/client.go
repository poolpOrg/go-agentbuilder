package client

import (
	"fmt"
	"net"
	"net/url"
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

type Client struct {
	remoteAddr string
	protocol   *protocol.Protocol
	exit       chan struct{}
}

func NewClient(address string) *Client {
	return &Client{
		remoteAddr: address,
		exit:       make(chan struct{}),
	}
}

func (c *Client) tcpConnect() (net.Conn, error) {
	location, err := url.Parse("ignore://" + c.remoteAddr)
	if err != nil {
		return nil, err
	}

	port := location.Port()
	if port == "" {
		port = "12457"
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", location.Hostname()+":"+port)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	return conn, err
}

func (c *Client) Run(handler func(*Session, <-chan protocol.Packet) error) error {
	conn, err := c.tcpConnect()
	if err != nil {
		return err
	}
	defer conn.Close()

	c.protocol = protocol.NewProtocol(conn)
	session := NewSession(conn)

	fmt.Printf("%s : connected\n", conn.RemoteAddr())
	for {
		if err = handler(session, session.protocol.Incoming()); err != nil {
			fmt.Printf("%s : error: %s\n", conn.RemoteAddr(), err)
		}
		break
	}
	fmt.Printf("%s : disconnected\n", conn.RemoteAddr())
	return err
}

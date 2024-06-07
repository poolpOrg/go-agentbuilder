package main

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/poolpOrg/go-agentbuilder/client"
	"github.com/poolpOrg/go-agentbuilder/protocol"
	"github.com/poolpOrg/go-agentbuilder/server"
)

type ReqPing struct {
	Timestamp time.Time
}
type ResPing struct {
	Timestamp time.Time
}

func serverHandler(s *server.Session, incoming <-chan protocol.Packet) error {
	now := time.Now()
	fmt.Printf("[%s] <- ping request at %s\n", s.RemoteAddr(), now)
	s.Query(ReqPing{Timestamp: now}, func(payload interface{}) error {
		switch payload := payload.(type) {
		case ResPing:
			recvTime := time.Now()
			fmt.Printf("[%s] -> ping reply at %s (latency: %s)\n", s.RemoteAddr(), recvTime, s.Latency())
		default:
			return fmt.Errorf("unsupported payload type %T", payload)
		}
		return nil
	})

	for packet := range incoming {
		switch payload := packet.Payload.(type) {
		//case agent.ResPing:
		//	fmt.Printf("[%s] -> ping reply at %s\n", s.conn.RemoteAddr(), payload.Timestamp)
		//	s.protocol.Response(packet.Uuid, agent.ResPing{})
		default:
			return fmt.Errorf("unsupported payload type %T", payload)
		}
	}
	return nil
}

func clientHandler(s *client.Session, incoming <-chan protocol.Packet) error {
	for packet := range incoming {
		switch payload := packet.Payload.(type) {
		case ReqPing:
			now := time.Now()
			fmt.Printf("[%s] -> ping request at %s\n", s.RemoteAddr(), payload.Timestamp)
			if err := packet.Response(ResPing{Timestamp: now}); err != nil {
				return err
			}
			fmt.Printf("[%s] <- ping reply at %s (latency: %s)\n", s.RemoteAddr(), now, s.Latency())

		//case agent.ResPing:
		//	fmt.Printf("[%s] -> ping reply at %s\n", s.conn.RemoteAddr(), payload.Timestamp)
		//	s.protocol.Response(packet.Uuid, agent.ResPing{})
		default:
			return fmt.Errorf("unsupported payload type %T", payload)
		}
	}
	return nil
}

func main() {
	var connectAddr string
	var listenAddr string

	flag.StringVar(&connectAddr, "connect", "", "connect address")
	flag.StringVar(&listenAddr, "listen", "", "listen address")
	flag.Parse()

	protocol.Register(ReqPing{})
	protocol.Register(ResPing{})

	wg := sync.WaitGroup{}

	if listenAddr != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s := server.NewServer(listenAddr)
			s.ListenAndServe(serverHandler)
		}()
	}

	if connectAddr != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := client.NewClient(connectAddr)
			c.Run(clientHandler)
		}()
	}

	wg.Wait()
}

package protocol

import (
	"encoding/gob"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Protocol struct {
	conn net.Conn

	encoder *gob.Encoder
	decoder *gob.Decoder

	inflightRequests map[uuid.UUID]chan Packet
	inflightMutex    sync.Mutex

	queryResponses chan Packet

	latency time.Duration
}

func NewProtocol(conn net.Conn) *Protocol {
	return &Protocol{
		conn: conn,

		encoder: gob.NewEncoder(conn),
		decoder: gob.NewDecoder(conn),

		inflightRequests: make(map[uuid.UUID]chan Packet),
		queryResponses:   make(chan Packet),
	}
}

func (p *Protocol) LocalAddr() net.Addr {
	return p.conn.LocalAddr()
}
func (p *Protocol) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}

func (p *Protocol) Latency() time.Duration {
	return p.latency
}

func (p *Protocol) Incoming() <-chan Packet {
	pchan := make(chan Packet)
	go func() {
		for m := range p.queryResponses {
			p.inflightMutex.Lock()
			notify := p.inflightRequests[uuid.MustParse(m.UUID)]
			p.inflightMutex.Unlock()
			notify <- m
		}
	}()

	go func() {
		defer p.conn.Close()
		defer close(pchan)
		for {
			result := Packet{}
			err := p.decoder.Decode(&result)
			if err != nil {
				return
			}

			result.protocol = p
			p.inflightMutex.Lock()
			_, ok := p.inflightRequests[uuid.MustParse(result.UUID)]
			p.inflightMutex.Unlock()
			if ok {
				p.queryResponses <- result
			} else {
				pchan <- result
			}
		}
	}()

	return pchan
}

func (p *Protocol) Request(payload interface{}) error {
	return p.encoder.Encode(newPacket(p, uuid.NewString(), payload))
}

func (p *Protocol) Query(payload interface{}, cb func(interface{}) error) error {
	packetID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	packet := newPacket(p, packetID.String(), payload)

	notify := make(chan Packet)
	p.inflightMutex.Lock()
	p.inflightRequests[packetID] = notify
	p.inflightMutex.Unlock()

	err = p.encoder.Encode(&packet)
	if err != nil {
		return err
	}

	now := time.Now()
	result := <-notify
	p.latency = time.Since(now)

	p.inflightMutex.Lock()
	delete(p.inflightRequests, packetID)
	p.inflightMutex.Unlock()
	close(notify)

	return cb(result.Payload)
}

package protocol

import "encoding/gob"

type Packet struct {
	protocol *Protocol
	UUID     string
	Payload  interface{}
}

func newPacket(protocol *Protocol, UUID string, payload interface{}) Packet {
	return Packet{
		protocol: protocol,
		UUID:     UUID,
		Payload:  payload,
	}
}
func (p Packet) Response(payload interface{}) error {
	return p.protocol.encoder.Encode(&Packet{
		protocol: p.protocol,
		UUID:     p.UUID,
		Payload:  payload,
	})
}

func init() {
	gob.Register(Packet{})
}

func Register(payload interface{}) {
	gob.Register(payload)
}

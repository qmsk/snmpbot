package client

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"io"
	"net"
)

var EOF = io.EOF

// key used to match send/recv IOs by the Engine
type ioKey struct {
	id        int
	community string
	addr      string
}

func (key ioKey) String() string {
	return fmt.Sprintf("%v@%v[%d]", key.community, key.addr, key.id)
}

type IO struct {
	Addr    net.Addr
	Packet  snmp.Packet
	PDUType snmp.PDUType
	PDU     snmp.PDU
}

func (io IO) key() ioKey {
	return ioKey{
		id:        io.PDU.GetRequestID(),
		community: string(io.Packet.Community),
		addr:      io.Addr.String(),
	}
}

type Transport interface {
	Resolve(addr string) (net.Addr, error)
	Send(IO) error
	Recv() (IO, error)
	Close() error
}

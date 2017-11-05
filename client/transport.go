package client

import (
	snmp "github.com/qmsk/snmpbot/snmp_new"
	"io"
	"net"
)

var EOF = io.EOF

type IO struct {
	Addr    net.Addr
	Packet  snmp.Packet
	PDUType snmp.PDUType
	PDU     snmp.PDU
}

type Transport interface {
	Send(IO) error
	Recv() (IO, error)
	Close() error
}

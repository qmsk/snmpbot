package snmp

import (
	"encoding/asn1"
	//"github.com/geoffgarside/ber"
	"net"
	"time"
)

type OID = asn1.ObjectIdentifier

const (
	UDP_SIZE = 16 * 1024
)

type Version int

const (
	SNMPv1  Version = 0
	SNMPv2c Version = 1
)

type PDUType int // context-specific

const (
	GetRequestType     PDUType = 0
	GetNextRequestType PDUType = 1
	GetResponseType    PDUType = 2
	SetRequestType     PDUType = 3
	TrapV1Type         PDUType = 4
	GetBulkRequestType PDUType = 5
	InformRequestType  PDUType = 6
	TrapV2Type         PDUType = 7
	ReportType         PDUType = 8
)

type GenericTrap int

const (
	TrapColdStart             GenericTrap = 0
	TrapWarmStart             GenericTrap = 1
	TrapLinkDown              GenericTrap = 2
	TrapLinkUp                GenericTrap = 3
	TrapAuthenticationFailure GenericTrap = 4
	TrapEgpNeighborLoss       GenericTrap = 5
	TrapEnterpriseSpecific    GenericTrap = 6
)

type ValueError int // context-specific NULLs in VarBind.Value

const (
	NoSuchObjectError   ValueError = 0
	NoSuchInstanceError ValueError = 1
	EndOfMibViewError   ValueError = 2
)

type Packet struct {
	Version   Version
	Community []byte
	PDU       asn1.RawValue
}

type PDU struct {
	RequestID   int
	ErrorStatus int
	ErrorIndex  int
	VarBinds    []VarBind
}

// SNMPv1 Trap-PDU
type TrapPDU struct {
	Enterprise   OID
	AgentAddr    net.IP // []byte
	GenericTrap  GenericTrap
	SpecificTrap int
	TimeStamp    time.Duration // int64
	VarBinds     []VarBind
}

type VarBind struct {
	Name  OID
	Value interface{}
}

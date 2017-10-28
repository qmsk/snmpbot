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

type ErrorValue int // context-specific NULLs in VarBind.Value

const (
	NoSuchObjectValue   ErrorValue = 0
	NoSuchInstanceValue ErrorValue = 1
	EndOfMibViewValue   ErrorValue = 2
)

type ApplicationValueType int // application-specific values in VarBind.Value

const (
	IPAddressType   ApplicationValueType = 0
	Counter32Type   ApplicationValueType = 1
	Gauge32Type     ApplicationValueType = 2
	TimeTicks32Type ApplicationValueType = 3
	OpaqueType      ApplicationValueType = 4
)

type Packet struct {
	Version   Version
	Community []byte
	RawPDU    asn1.RawValue
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
	Name     OID
	RawValue asn1.RawValue
}

type IPAddress net.IP
type Counter32 uint32
type Gauge32 uint32
type TimeTicks32 uint32 // duration of 1/100 s
type Opaque []byte

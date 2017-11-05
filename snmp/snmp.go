package snmp

import (
	"encoding/asn1"
	"fmt"
	"net"
	"strings"
	"time"
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

func (pduType PDUType) String() string {
	switch pduType {
	case GetRequestType:
		return "GetRequest"
	case GetNextRequestType:
		return "GetNextRequest"
	case GetResponseType:
		return "GetResponse"
	case SetRequestType:
		return "SetRequest"
	case TrapV1Type:
		return "TrapV1"
	case GetBulkRequestType:
		return "GetBulkRequest"
	case InformRequestType:
		return "InformRequest"
	case TrapV2Type:
		return "TrapV2"
	case ReportType:
		return "Report"
	default:
		return fmt.Sprintf("PDUType(%d)", pduType)
	}
}

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

type ErrorStatus int

const (
	Success         ErrorStatus = 0
	TooBigError     ErrorStatus = 1
	NoSuchNameError ErrorStatus = 2
	BadValueError   ErrorStatus = 3
	ReadOnlyError   ErrorStatus = 4
	GenericError    ErrorStatus = 5
)

func (err ErrorStatus) String() string {
	switch err {
	case Success:
		return "Success"
	case TooBigError:
		return "TooBig"
	case NoSuchNameError:
		return "NoSuchName"
	case BadValueError:
		return "BadValue"
	case ReadOnlyError:
		return "ReadOnly"
	case GenericError:
		return "GenericError"
	default:
		return fmt.Sprintf("ErrorStatus(%d)", err)
	}
}

func (err ErrorStatus) Error() string {
	return fmt.Sprintf("SNMP PDU Error: %s", err.String())
}

type ErrorValue int // context-specific NULLs in VarBind.Value

const (
	NoSuchObjectValue   ErrorValue = 0
	NoSuchInstanceValue ErrorValue = 1
	EndOfMibViewValue   ErrorValue = 2
)

func (err ErrorValue) String() string {
	switch err {
	case NoSuchObjectValue:
		return "NoSuchObject"
	case NoSuchInstanceValue:
		return "NoSuchInstance"
	case EndOfMibViewValue:
		return "EndOfMibView"
	default:
		return fmt.Sprintf("ErrorValue(%d)", err)
	}
}

func (err ErrorValue) Error() string {
	return fmt.Sprintf("SNMP VarBind Error: %s", err.String())
}

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
	ErrorStatus ErrorStatus
	ErrorIndex  int
	VarBinds    []VarBind
}

func (pdu PDU) String() string {
	if pdu.ErrorStatus != 0 {
		return fmt.Sprintf("!%v", pdu.ErrorStatus)
	}

	var varBinds = make([]string, len(pdu.VarBinds))

	for i, varBind := range pdu.VarBinds {
		varBinds[i] = varBind.String()
	}

	return strings.Join(varBinds, ", ")
}

// SNMPv1 Trap-PDU
type TrapPDU struct {
	Enterprise   asn1.ObjectIdentifier
	AgentAddr    net.IP // []byte
	GenericTrap  GenericTrap
	SpecificTrap int
	TimeStamp    time.Duration // int64
	VarBinds     []VarBind
}

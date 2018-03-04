package snmp

import (
	"encoding/asn1"
	"fmt"
	"strings"
)

type GenericPDU struct {
	RequestID   int
	ErrorStatus ErrorStatus
	ErrorIndex  int
	VarBinds    []VarBind
}

func (pdu *GenericPDU) unpack(raw asn1.RawValue) error {
	return unpack(raw, pdu)
}

func (pdu GenericPDU) GetRequestID() int {
	return pdu.RequestID
}

func (pdu GenericPDU) String() string {
	if pdu.ErrorStatus != 0 {
		return fmt.Sprintf("!%v", pdu.ErrorStatus)
	}

	var varBinds = make([]string, len(pdu.VarBinds))

	for i, varBind := range pdu.VarBinds {
		varBinds[i] = varBind.String()
	}

	return strings.Join(varBinds, ", ")
}

func (pdu GenericPDU) ErrorVarBind() VarBind {
	if pdu.ErrorIndex < len(pdu.VarBinds) {
		return pdu.VarBinds[pdu.ErrorIndex]
	} else {
		return VarBind{}
	}
}

func (pdu GenericPDU) Pack(pduType PDUType) (asn1.RawValue, error) {
	return packSequence(asn1.ClassContextSpecific, int(pduType),
		pdu.RequestID,
		pdu.ErrorStatus,
		pdu.ErrorIndex,
		pdu.VarBinds,
	)
}

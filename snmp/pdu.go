package snmp

import (
	"encoding/asn1"
	"fmt"
	"strings"
)

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

func (pdu PDU) ErrorVarBind() VarBind {
	if pdu.ErrorIndex < len(pdu.VarBinds) {
		return pdu.VarBinds[pdu.ErrorIndex]
	} else {
		return VarBind{}
	}
}

func (pdu PDU) Pack(pduType PDUType) (asn1.RawValue, error) {
	return packSequence(asn1.ClassContextSpecific, int(pduType),
		pdu.RequestID,
		pdu.ErrorStatus,
		pdu.ErrorIndex,
		pdu.VarBinds,
	)
}

func (pdu *PDU) Unpack(raw asn1.RawValue) error {
	if raw.Class != asn1.ClassContextSpecific {
		return fmt.Errorf("unexpected PDU: ASN.1 class=%d tag=%d", raw.Class, raw.Tag)
	}

	return unpack(raw, pdu)
}

package snmp

import (
	"encoding/asn1"
)

// Very similar to the PDU type, but the error fields are replaced by parameters
type BulkPDU struct {
	RequestID      int
	NonRepeaters   int
	MaxRepetitions int
	VarBinds       []VarBind
}

func (pdu *BulkPDU) unpack(raw asn1.RawValue) error {
	return unpack(raw, pdu)
}

func (pdu BulkPDU) GetRequestID() int {
	return pdu.RequestID
}

func (pdu BulkPDU) Pack(pduType PDUType) (asn1.RawValue, error) {
	return packSequence(asn1.ClassContextSpecific, int(pduType),
		pdu.RequestID,
		pdu.NonRepeaters,
		pdu.MaxRepetitions,
		pdu.VarBinds,
	)
}

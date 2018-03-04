package snmp

import (
	"encoding/asn1"
	"fmt"
)

// Very similar to the PDU type, but the error fields are replaced by parameters
type BulkPDU struct {
	RequestID      int
	NonRepeaters   int
	MaxRepetitions int
	VarBinds       []VarBind
}

func (pdu BulkPDU) Pack(pduType PDUType) (asn1.RawValue, error) {
	return packSequence(asn1.ClassContextSpecific, int(pduType),
		pdu.RequestID,
		pdu.NonRepeaters,
		pdu.MaxRepetitions,
		pdu.VarBinds,
	)
}

func (pdu *BulkPDU) Unpack(raw asn1.RawValue) error {
	if raw.Class != asn1.ClassContextSpecific {
		return fmt.Errorf("unexpected PDU: ASN.1 class=%d tag=%d", raw.Class, raw.Tag)
	}

	return unpack(raw, pdu)
}

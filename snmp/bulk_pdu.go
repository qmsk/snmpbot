package snmp

import (
	"encoding/asn1"
	"fmt"
	"strings"
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
func (pdu BulkPDU) SetRequestID(id int) {
	pdu.RequestID = id
}

func (pdu BulkPDU) String() string {
	var scalars []string
	var entries []string

	for i, varBind := range pdu.VarBinds {
		if i < pdu.NonRepeaters {
			scalars = append(scalars, varBind.String())
		} else {
			entries = append(entries, varBind.String())
		}
	}

	return fmt.Sprintf("[%v] + %dx[%v]", strings.Join(scalars, ", "), pdu.MaxRepetitions, strings.Join(entries, ", "))
}

func (pdu BulkPDU) GetError() PDUError {
	return PDUError{}
}

func (pdu BulkPDU) Pack(pduType PDUType) (asn1.RawValue, error) {
	return packSequence(asn1.ClassContextSpecific, int(pduType),
		pdu.RequestID,
		pdu.NonRepeaters,
		pdu.MaxRepetitions,
		pdu.VarBinds,
	)
}

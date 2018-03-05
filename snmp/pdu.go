package snmp

import (
	"encoding/asn1"
	"fmt"
)

type PDUError struct {
	ErrorStatus ErrorStatus
	VarBind     VarBind
}

type PDU interface {
	GetRequestID() int

	GetError() PDUError

	Pack(PDUType) (asn1.RawValue, error)
}

func UnpackPDU(raw asn1.RawValue) (PDUType, PDU, error) {
	var pduType = PDUType(raw.Tag)

	if raw.Class != asn1.ClassContextSpecific {
		return pduType, nil, fmt.Errorf("unexpected PDU: ASN.1 class=%d tag=%d", raw.Class, raw.Tag)
	}

	switch pduType {
	case GetRequestType, GetNextRequestType, GetResponseType, SetRequestType:
		var pdu GenericPDU

		err := pdu.unpack(raw)

		return pduType, pdu, err

	case GetBulkRequestType:
		var pdu BulkPDU

		err := pdu.unpack(raw)

		return pduType, pdu, err

	default:
		return pduType, nil, fmt.Errorf("Unknown PDUType=%v", pduType)
	}
}

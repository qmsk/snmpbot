package snmp

import (
	"encoding/asn1"
	"fmt"
)

type PDUMeta struct {
	PDUType   PDUType
	RequestID int
}

type PDUError struct {
	ErrorStatus ErrorStatus
	VarBind     VarBind
}

type PDU interface {
	GetRequestID() int

	GetError() PDUError

	Pack(PDUMeta) (asn1.RawValue, error)
}

func UnpackPDU(raw asn1.RawValue) (PDUMeta, PDU, error) {
	var pduType = PDUType(raw.Tag)

	if raw.Class != asn1.ClassContextSpecific {
		return PDUMeta{PDUType: pduType}, nil, fmt.Errorf("unexpected PDU: ASN.1 class=%d tag=%d", raw.Class, raw.Tag)
	}

	switch pduType {
	case GetRequestType, GetNextRequestType, GetResponseType, SetRequestType:
		var pdu GenericPDU

		err := pdu.unpack(raw)

		return PDUMeta{pduType, pdu.RequestID}, pdu, err

	case GetBulkRequestType:
		var pdu BulkPDU

		err := pdu.unpack(raw)

		return PDUMeta{pduType, pdu.RequestID}, pdu, err

	default:
		return PDUMeta{PDUType: pduType}, nil, fmt.Errorf("Unknown PDUType=%v", pduType)
	}
}

package snmp

import (
	"encoding/asn1"
	"fmt"
	"github.com/geoffgarside/ber"
)

func unpack(raw asn1.RawValue, value interface{}) error {
	var params string

	switch raw.Class {
	case asn1.ClassUniversal:

	case asn1.ClassContextSpecific:
		params = fmt.Sprintf("tag:%d", raw.Tag)
	case asn1.ClassApplication:
		params = fmt.Sprintf("application,tag:%d", raw.Tag)
	default:
		return fmt.Errorf("unable to unpack raw value with class=%d", raw.Class)
	}

	if _, err := ber.UnmarshalWithParams(raw.FullBytes, value, params); err != nil {
		return err
	} else {
		// ignore trailing bytes
	}

	return nil
}

func (packet *Packet) Unmarshal(buf []byte) error {
	if _, err := ber.Unmarshal(buf, packet); err != nil {
		return err
	} else {
		// ignore trailing bytes
	}

	if packet.RawPDU.Class != asn1.ClassContextSpecific {
		return fmt.Errorf("unexpected PDU: ASN.1 class %d", packet.RawPDU.Class)
	}

	return nil
}

func (packet *Packet) PDUType() PDUType {
	// assuming packet.PDU.Class == asn1.ClassContextSpecific
	return PDUType(packet.RawPDU.Tag)
}

func (pdu *PDU) Unpack(raw asn1.RawValue) error {
	if raw.Class != asn1.ClassContextSpecific {
		return fmt.Errorf("unexpected PDU: ASN.1 class=%d tag=%d", raw.Class, raw.Tag)
	}

	return unpack(raw, pdu)
}

package snmp

import (
	"encoding/asn1"
	"fmt"
	"github.com/geoffgarside/ber"
)

func (packet *Packet) unmarshal(buf []byte) error {
	if _, err := ber.Unmarshal(buf, packet); err != nil {
		return err
	} else { // ignore trailing bytes
		return nil
	}

	if packet.PDU.Class != asn1.ClassContextSpecific {
		return fmt.Errorf("unexpected PDU: ASN.1 class %d", packet.PDU.Class)
	}

	return nil
}

func (packet *Packet) PDUType() PDUType {
	// assuming packet.PDU.Class == asn1.ClassContextSpecific
	return PDUType(packet.PDU.Tag)
}

func (pdu *PDU) unpack(raw asn1.RawValue) error {
	if raw.Class != asn1.ClassContextSpecific {
		return fmt.Errorf("unexpected PDU: ASN.1 class=%d tag=%d", raw.Class, raw.Tag)
	}

	var params = fmt.Sprintf("tag:%d", raw.Tag)

	if _, err := ber.UnmarshalWithParams(raw.FullBytes, pdu, params); err != nil {
		return err
	} else { // ignore trailing bytes
		return nil
	}
}

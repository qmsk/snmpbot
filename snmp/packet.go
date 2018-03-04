package snmp

import (
	"encoding/asn1"
	"fmt"
)

type Packet struct {
	Version   Version
	Community []byte
	RawPDU    asn1.RawValue
}

func (packet Packet) Marshal() ([]byte, error) {
	return marshal(packet)
}

func (packet *Packet) Unmarshal(buf []byte) error {
	if err := unmarshal(buf, packet); err != nil {
		return err
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

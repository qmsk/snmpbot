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

func (packet *Packet) UnpackPDU() (PDUType, PDU, error) {
	return UnpackPDU(packet.RawPDU)
}

func (packet *Packet) PackPDU(pduType PDUType, pdu PDU) error {
	if rawPDU, err := pdu.Pack(pduType); err != nil {
		return err
	} else {
		packet.RawPDU = rawPDU
	}

	return nil
}

func (packet *Packet) Marshal() ([]byte, error) {
	return marshal(*packet)
}

package snmp

import (
	"encoding/asn1"
)

// XXX: workaround lack of asn1.MarshalWithParameters
func pack(cls int, tag int, value interface{}) (asn1.RawValue, error) {
	var raw asn1.RawValue

	if value == nil {
		raw = asn1.NullRawValue
	} else if buf, err := asn1.Marshal(value); err != nil {
		return raw, err
	} else if _, err := asn1.Unmarshal(buf, &raw); err != nil {
		return raw, err
	}

	if cls != asn1.ClassUniversal {
		raw.Class = cls
		raw.Tag = tag
		raw.FullBytes = nil
	}

	return raw, nil
}

func packSequence(cls int, tag int, values ...interface{}) (asn1.RawValue, error) {
	var raw = asn1.RawValue{Class: cls, Tag: tag, IsCompound: true}

	for _, value := range values {
		if value == nil {
			raw.Bytes = append(raw.Bytes, asn1.NullBytes...)
		} else if bytes, err := asn1.Marshal(value); err != nil {
			return raw, err
		} else {
			raw.Bytes = append(raw.Bytes, bytes...)
		}
	}

	return raw, nil
}

func marshalSequence(cls int, tag int, values ...interface{}) ([]byte, error) {
	if raw, err := packSequence(cls, tag, values...); err != nil {
		return nil, err
	} else {
		return asn1.Marshal(raw)
	}
}

func (packet Packet) Marshal() ([]byte, error) {
	return asn1.Marshal(packet)
}

func (pdu PDU) Pack(pduType PDUType) (asn1.RawValue, error) {
	return packSequence(asn1.ClassContextSpecific, int(pduType),
		pdu.RequestID,
		pdu.ErrorStatus,
		pdu.ErrorIndex,
		pdu.VarBinds,
	)
}

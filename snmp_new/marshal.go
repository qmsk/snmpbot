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
	for i, varBind := range pdu.VarBinds {
		if varBind.RawValue.Class == 0 && varBind.RawValue.Tag == 0 {
			pdu.VarBinds[i].RawValue = asn1.NullRawValue
		}
	}

	return packSequence(asn1.ClassContextSpecific, int(pduType),
		pdu.RequestID,
		pdu.ErrorStatus,
		pdu.ErrorIndex,
		pdu.VarBinds,
	)
}

func (varBind *VarBind) Set(genericValue interface{}) error {
	switch value := genericValue.(type) {
	case nil:
		varBind.SetNull()
	case ErrorValue:
		return varBind.SetError(value)
	case IPAddress:
		return varBind.setApplication(IPAddressType, value)
	case Counter32:
		return varBind.setApplication(Counter32Type, int(value))
	case Gauge32:
		return varBind.setApplication(Gauge32Type, int(value))
	case TimeTicks32:
		return varBind.setApplication(TimeTicks32Type, int(value))
	case Opaque:
		return varBind.setApplication(OpaqueType, value)
	default:
		if rawValue, err := pack(asn1.ClassUniversal, 0, value); err != nil {
			return err
		} else {
			varBind.RawValue = rawValue
		}
	}

	return nil
}

func (varBind *VarBind) SetNull() {
	varBind.RawValue = asn1.NullRawValue
}

func (varBind *VarBind) SetError(errorValue ErrorValue) error {
	if rawValue, err := pack(asn1.ClassContextSpecific, int(errorValue), nil); err != nil {
		return err
	} else {
		varBind.RawValue = rawValue
	}

	return nil
}

func (varBind *VarBind) setApplication(tag ApplicationValueType, value interface{}) error {
	if rawValue, err := pack(asn1.ClassApplication, int(tag), value); err != nil {
		return err
	} else {
		varBind.RawValue = rawValue
	}

	return nil
}

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

func (packet *Packet) unmarshal(buf []byte) error {
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

func (pdu *PDU) unpack(raw asn1.RawValue) error {
	if raw.Class != asn1.ClassContextSpecific {
		return fmt.Errorf("unexpected PDU: ASN.1 class=%d tag=%d", raw.Class, raw.Tag)
	}

	return unpack(raw, pdu)
}

func (varBind VarBind) Value() (interface{}, error) {
	switch varBind.RawValue.Class {
	case asn1.ClassUniversal:
		var value interface{}

		return value, unpack(varBind.RawValue, &value)

	case asn1.ClassApplication:
		switch ApplicationValueType(varBind.RawValue.Tag) {
		case IPAddressType:
			var value IPAddress

			return value, unpack(varBind.RawValue, &value)

		case Counter32Type:
			var value int

			if err := unpack(varBind.RawValue, &value); err != nil {
				return nil, err
			} else {
				return Counter32(value), nil
			}

		case Gauge32Type:
			var value int

			if err := unpack(varBind.RawValue, &value); err != nil {
				return nil, err
			} else {
				return Gauge32(value), nil
			}

		case TimeTicks32Type:
			var value int

			if err := unpack(varBind.RawValue, &value); err != nil {
				return nil, err
			} else {
				return TimeTicks32(value), nil
			}

		case OpaqueType:
			var value Opaque

			return value, unpack(varBind.RawValue, &value)

		default:
			return nil, fmt.Errorf("Unkown varbind value application tag=%d", varBind.RawValue.Tag)
		}

	case asn1.ClassContextSpecific:
		switch ErrorValue(varBind.RawValue.Tag) {
		case NoSuchObjectValue:
			return NoSuchObjectValue, nil
		case NoSuchInstanceValue:
			return NoSuchInstanceValue, nil
		case EndOfMibViewValue:
			return EndOfMibViewValue, nil
		default:
			return nil, fmt.Errorf("Unkown varbind value context-specific tag=%d", varBind.RawValue.Tag)
		}

	default:
		return nil, fmt.Errorf("Unkown varbind value class=%d", varBind.RawValue.Class)
	}
}

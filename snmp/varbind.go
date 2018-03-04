package snmp

import (
	"encoding/asn1"
	"fmt"
)

type IPAddress [4]uint8
type Counter32 uint32
type Gauge32 uint32
type TimeTicks32 uint32 // duration of 1/100 s
type Opaque []byte
type Counter64 uint64

// panics if unable to pack value
func MakeVarBind(oid OID, value interface{}) VarBind {
	var varBind = VarBind{
		Name: asn1.ObjectIdentifier(oid),
	}

	if err := varBind.Set(value); err != nil {
		panic(err)
	}

	return varBind
}

type VarBind struct {
	Name     asn1.ObjectIdentifier
	RawValue asn1.RawValue
}

func (varBind VarBind) String() string {
	if len(varBind.Name) == 0 {
		return fmt.Sprintf(".")
	} else if value, err := varBind.Value(); err != nil {
		return fmt.Sprintf("!%v", varBind.Name)
	} else if value != nil {
		return fmt.Sprintf("%v=%v", varBind.Name, value)
	} else {
		return fmt.Sprintf("%v", varBind.Name)
	}
}

func (varBind VarBind) OID() OID {
	return OID(varBind.Name)
}

// Return ErrorValue if exists, otherwise nil
func (varBind VarBind) ErrorValue() error {
	if varBind.RawValue.Class == asn1.ClassContextSpecific {
		return ErrorValue(varBind.RawValue.Tag)
	}

	return nil
}

func (varBind VarBind) Value() (interface{}, error) {
	switch varBind.RawValue.Class {
	case asn1.ClassUniversal:
		if varBind.RawValue.Tag == asn1.TagNull {
			return nil, nil
		} else {
			var value interface{}

			return value, unpack(varBind.RawValue, &value)
		}

	case asn1.ClassApplication:
		switch ApplicationValueType(varBind.RawValue.Tag) {
		case IPAddressType:
			var value []byte

			if err := unpack(varBind.RawValue, &value); err != nil {
				return nil, err
			} else if len(value) != 4 {
				return nil, fmt.Errorf("Invalid IPAddress value: %#v", value)
			} else {
				var ipAddress IPAddress

				for i := 0; i < 4; i++ {
					ipAddress[i] = uint8(value[i])
				}

				return ipAddress, nil
			}

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

		case Counter64Type:
			var value int64 // XXX: no support for uint64?

			if err := unpack(varBind.RawValue, &value); err != nil {
				return nil, err
			} else {
				return Counter64(value), nil
			}

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

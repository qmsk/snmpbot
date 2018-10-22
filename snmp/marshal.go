package snmp

import (
	"encoding/asn1"
	"fmt"
)

// pack value with custom class/tag, returning a RawValue for further use in sequences/etc.
// returns RawValue with FullBytes set. The .Bytes will always be nil.
func pack(cls int, tag int, value interface{}) (asn1.RawValue, error) {
	var params string
	var raw = asn1.RawValue{
		Class: cls,
		Tag:   tag,
	}

	// asn.MarshalWithParams does not allow marshalling non-raw nil values
	if value == nil {
		raw.Bytes = []byte{} // explicit empty value for nil

		if bytes, err := asn1.Marshal(raw); err != nil {
			return raw, err
		} else {
			raw.FullBytes = bytes
		}
	} else {
		switch raw.Class {
		case asn1.ClassUniversal:

		case asn1.ClassContextSpecific:
			params = fmt.Sprintf("tag:%d", raw.Tag)
		case asn1.ClassApplication:
			params = fmt.Sprintf("application,tag:%d", raw.Tag)
		default:
			return raw, fmt.Errorf("unable to unpack raw value with class=%d", raw.Class)
		}

		if bytes, err := asn1.MarshalWithParams(value, params); err != nil {
			return raw, err
		} else {
			raw.FullBytes = bytes
		}
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

func marshal(obj interface{}) ([]byte, error) {
	return asn1.Marshal(obj)
}

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

func unmarshal(data []byte, obj interface{}) error {
	if _, err := ber.Unmarshal(data, obj); err != nil {
		return err
	} else {
		// XXX: ignore trailing bytes
	}

	return nil
}

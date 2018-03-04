package snmp

import (
	"fmt"
	"strconv"
	"strings"
)

type OID []int

// panic on ParseOID errors
func MustParseOID(str string) OID {
	if oid, err := ParseOID(str); err != nil {
		panic(err)
	} else {
		return oid
	}
}

func ParseOID(str string) (OID, error) {
	if str == "" {
		return nil, nil
	} else if str == "." {
		return OID{}, nil
	} else if str[0] != '.' {
		return nil, fmt.Errorf("Invalid OID: does not start with .")
	} else {
		str = str[1:]
	}

	var parts = strings.Split(str, ".")
	var oid = make(OID, len(parts))

	for i, part := range parts {
		if id, err := strconv.Atoi(part); err != nil {
			return nil, fmt.Errorf("Invalid OID part: %v", part)
		} else {
			oid[i] = id
		}
	}

	return oid, nil
}

func (oid OID) String() (str string) {
	if oid == nil {
		return ""
	}
	if len(oid) == 0 {
		return "."
	}

	for _, id := range oid {
		str += fmt.Sprintf(".%d", id)
	}
	return str
}

func (oid OID) Copy() OID {
	var copy OID

	return append(copy, oid...)
}

// Extend this OID with the given ids, returning the new, more-specific, OID.
func (oid OID) Extend(ids ...int) OID {
	return append(oid.Copy(), ids...)
}

// Compare two OIDs for equality
func (oid OID) Equals(other OID) bool {
	if len(oid) != len(other) {
		return false
	}
	for i := range oid {
		if oid[i] != other[i] {
			return false
		}
	}
	return true
}

// Test if the given OID is a more-specific of this OID, returning the extended part if so.
// Returns {} if the OIDs are an exact match
// Returns nil if the OIDs do not match
func (oid OID) Index(other OID) []int {
	if len(other) < len(oid) {
		return nil
	}

	for i := range oid {
		if oid[i] != other[i] {
			return nil
		}
	}

	if len(other) == len(oid) {
		return OID{}
	} else {
		return other[len(oid):]
	}
}

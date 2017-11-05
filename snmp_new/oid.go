package snmp

import (
	"fmt"
	"strconv"
	"strings"
)

type OID []int

func ParseOID(str string) (oid OID) {
	if len(str) > 0 && str[0] == '.' {
		str = str[1:]
	}

	if str == "" {
		return OID{}
	}

	parts := strings.Split(str, ".")

	for _, part := range parts {
		if id, err := strconv.Atoi(part); err != nil {
			panic(err)
		} else {
			oid = append(oid, id)
		}
	}
	return
}

func (oid OID) String() (str string) {
	if oid == nil {
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
// Returns {0} if the OIDs are an exact match
// Returns nil if the OIDs do not match
func (oid OID) Index(other OID) (index OID) {
	if len(other) < len(oid) {
		return nil
	}

	for i := range oid {
		if oid[i] != other[i] {
			return nil
		}
	}

	if len(other) == len(oid) {
		return OID{0}
	} else {
		return other[len(oid):]
	}
}

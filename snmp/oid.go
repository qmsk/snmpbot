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

func (self OID) String() (str string) {
	if self == nil {
		return "."
	}

	for _, id := range self {
		str += fmt.Sprintf(".%d", id)
	}
	return str
}

func (self OID) Copy() OID {
	var oid OID

	oid = append(oid, self...)

	return oid
}

// Extend this OID with the given ids, returning the new, more-specific, OID.
func (self OID) define(ids ...int) OID {
	return append(self.Copy(), ids...)
}

// Compare two OIDs for equality
func (self OID) Equals(oid OID) bool {
	if len(self) != len(oid) {
		return false
	}
	for i := range self {
		if self[i] != oid[i] {
			return false
		}
	}
	return true
}

// Test if the given OID is a more-specific of this OID, returning the extended part if so.
// Returns {0} if the OIDs are an exact match
// Returns nil if the OIDs do not match
func (self OID) Index(oid OID) (index OID) {
	if len(oid) < len(self) {
		return nil
	}

	for i := range self {
		if self[i] != oid[i] {
			return nil
		}
	}

	if len(oid) == len(self) {
		return OID{0}
	} else {
		return oid[len(self):]
	}
}

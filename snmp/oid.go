package snmp

import (
    "fmt"
    "strings"
    "strconv"
)

type OID []int

func parseOID(str string) (oid OID) {
    parts := strings.Split(str, ".")

    for index, part := range parts {
        if index == 0 && part == "" {
            continue
        }

        if id, err := strconv.Atoi(part); err != nil {
            panic(err)
        } else {
            oid = append(oid, id)
        }
    }
    return
}

func (self OID) String() (str string) {
    for _, id := range self {
        str = str + fmt.Sprintf(".%d", id)
    }
    return
}

// Extend this OID with the given ids, returning the new, more-specific, OID.
func (self OID) define(ids... int) (defineOid OID) {
    defineOid = append(defineOid, self...)
    defineOid = append(defineOid, ids...)

    return
}

// Compare two OIDs for equality
func (self OID) Match(oid OID) bool {
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

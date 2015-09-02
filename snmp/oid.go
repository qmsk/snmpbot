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
        if index == 0 {
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

func (self OID) Index() (oid OID, index int) {
    split := len(self) - 1

    oid = self[0:split]
    index = self[split]

    return
}

func (self OID) define(oid... int) (defineOid OID) {
    defineOid = append(defineOid, self...)
    defineOid = append(defineOid, oid...)

    return
}

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

/* MIB */
type MIB struct {
    OID
}

/* Tables */
func (self OID) defineTable(oid int) Table {
    return Table{OID: self.define(oid)}
}

type Table struct {
    OID
}

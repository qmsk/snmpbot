package snmp

import (
	"fmt"
	"sort"
	"strings"
)

func MakeOIDSet(oids ...OID) OIDSet {
	// minimize
	var oidSet = make(OIDSet)

	for _, oid := range oids {
		oidSet.Add(oid)
	}

	return oidSet
}

type OIDSet map[string]OID

func (oidSet OIDSet) String() string {
	var strs []string

	for _, oid := range oidSet {
		strs = append(strs, oid.String())
	}

	sort.Strings(strs)

	return "{" + strings.Join(strs, " ") + "}"
}

func (oidSet OIDSet) Get(oid OID) OID {
	var key = ""

	if matchOID, ok := oidSet["."]; ok {
		return matchOID
	}

	for _, x := range oid {
		key += fmt.Sprintf(".%d", x)

		if matchOID, ok := oidSet[key]; ok {
			return matchOID
		}
	}

	return nil
}

func (oidSet OIDSet) Add(oid OID) {
	if oid == nil {
		panic("add nil oid to set")
	}

	for oidKey, o := range oidSet {
		if idx := o.Index(oid); idx != nil {
			// set already contains OID covering this OID
			return
		} else if idx := oid.Index(o); idx != nil {
			// delete OID from set covered by this OID
			delete(oidSet, oidKey)
		}
	}

	oidSet[oid.String()] = oid
}

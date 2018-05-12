package snmp

func MakeOIDSet(oids ...OID) OIDSet {
	// minimize
	var oidSet OIDSet

	for _, oid := range oids {
		oidSet.Add(oid)
	}

	return oidSet
}

type OIDSet []OID

func (oidSet OIDSet) Get(oid OID) OID {
	for _, o := range oidSet {
		if idx := o.Index(oid); idx != nil {
			return o
		}
	}

	return nil
}

func (oidSet *OIDSet) Add(oid OID) {
	if oid == nil {
		panic("add nil oid to set")
	}

	for i, o := range *oidSet {
		if idx := o.Index(oid); idx != nil {
			// set already contains OID covering this OID
			return
		} else if idx := oid.Index(o); idx != nil {
			// delete OID from set covered by this OID
			*oidSet = append((*oidSet)[:i], (*oidSet)[i+1:]...)
		}
	}

	*oidSet = append(*oidSet, oid)
}

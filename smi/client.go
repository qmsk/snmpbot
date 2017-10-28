
// Get a scalar SNMP object, returning its value
// Returns nil if the object is not found
func (self *Client) GetObject(object *Object) (interface{}, error) {
	if varBinds, err := self.Get(object.OID.define(0)); err != nil {
		return nil, err
	} else {
		for _, varBind := range varBinds {
			oid := OID(varBind.Name)

			if index := object.Index(oid); index == nil {
				return nil, fmt.Errorf("response var-bind OID mismatch")
			} else if varBind.Value == NoSuchObjectError || Value == NoSuchInstanceError {
				return nil, nil
			} else if objectValue, err := object.ParseValue(varBind.Value); err != nil {
				return nil, err
			} else {
				return objectValue, nil
			}
		}

		return nil, nil
	}
}

// Probe for supported MIBS
func (self *Client) ProbeMIBs(handler func(*MIB)) error {
	for _, mib := range mibs {
		// TODO: probe...
		handler(mib)
	}

	return nil
}

// Probe for supported Objects in given MIB
func (self *Client) ProbeMIBObjects(mib *MIB, handler func(*Object)) error {
	for _, object := range mib.objects {
		if object.Table != nil {
			continue
		}

		// TODO: probe...
		handler(object)
	}

	return nil
}

// Probe for supported Tables in given MIB
func (self *Client) ProbeMIBTables(mib *MIB, handler func(*Table)) error {
	for _, table := range mib.tables {
		// TODO: probe...
		handler(table)
	}

	return nil
}
